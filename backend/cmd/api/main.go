package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/tuusuario/afs-challenge/config"
	repo "github.com/tuusuario/afs-challenge/internal/infrastructure/database/repositories"
	handlers "github.com/tuusuario/afs-challenge/internal/presentation/http/handlers"
	"github.com/tuusuario/afs-challenge/internal/presentation/http/routes"
	applogger "github.com/tuusuario/afs-challenge/pkg/logger"
	"github.com/tuusuario/afs-challenge/internal/usecases"
)

func main() {
	// ============================================
	// 1. Load Configuration
	// ============================================
	cfg := config.Load()
	applogger.Info("🔧 Configuration loaded")
	applogger.Info("Environment: " + cfg.Environment)
	if err := cfg.ValidateTiger(); err != nil {
		log.Fatalf("Tiger Cloud config error: %v", err)
	}

	// ============================================
	// 2. Initialize Database
	// ============================================
	db := initDatabase(cfg)
	defer db.Close()
	applogger.Info("✅ Database connected")



	// ============================================
	// 3. Initialize Infrastructure Layer
	// ============================================
	// TODO: Initialize Redis connection
	// redisClient := initRedis(cfg.RedisURL)
	// defer redisClient.Close()

	// ============================================
	// 4. Initialize Repositories (Infrastructure)
	// ============================================
	// migrationRepo := postgres.NewPostgresMigrationRepository(db)

	// ============================================
	// 5. Initialize Use Cases (Application Layer)
	// ============================================
	// orchestrateMigrationUC := usecases.NewOrchestrateMigrationUseCase(migrationRepo)

	// ============================================
	// 6. Initialize WebSocket Hub
	// ============================================
	hub := usecases.NewHub()
	go hub.Run()

	// ============================================
	// 7. Setup Fiber App
	// ============================================
	app := fiber.New(fiber.Config{
		AppName:               "AFS Challenge API",
		ServerHeader:          "",
		DisableStartupMessage: false,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           30 * time.Second,
	})

	// ============================================
	// 8. Global Middlewares
	// ============================================
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.Environment == "development",
	}))

	app.Use(fiberlogger.New(fiberlogger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} (${latency})\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: getAllowedOrigins(cfg.Environment),
		AllowMethods: "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// ============================================
	// 9. Setup Routes (Dependency Injection here)
	// ============================================
	// DI: repos + services + handlers
	taskRepo := repo.NewPostgresTaskRepository(db)
	taskSvc := usecases.NewTaskService(taskRepo)
	taskHandler := handlers.NewTaskHandler(taskSvc, hub)
	agentExecRepo := repo.NewPostgresAgentExecutionRepository(db)
	optRepo := repo.NewPostgresOptimizationRepository(db)
	benchRepo := repo.NewPostgresBenchmarkRepository(db)
	consRepo := repo.NewPostgresConsensusRepository(db)
	resultsHandler := handlers.NewResultsHandler(agentExecRepo, optRepo, benchRepo, consRepo, hub)
	routes.SetupRoutes(app, hub, taskHandler, resultsHandler)

	// ============================================
	// 10. Graceful Shutdown
	// ============================================
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		applogger.Info("🛑 Shutting down gracefully...")

		db.Close()
		// redisClient.Close()

		if err := app.Shutdown(); err != nil {
			applogger.Error("Error during shutdown", err)
		}

		applogger.Info("✅ Server stopped")
	}()

	// ============================================
	// 11. Start Server
	// ============================================
	applogger.Info("🚀 Server starting on port " + cfg.Port)
	applogger.Info("Environment: " + cfg.Environment)

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// ============================================
// Helper Functions
// ============================================

func getAllowedOrigins(env string) string {
	if env == "production" {
		return "https://yourdomain.com"
	}
	return "*"
}

func initDatabase(cfg *config.Config) *sqlx.DB {
	dsn := cfg.DatabaseURL
	if dsn == "" {
		log.Fatal("database DSN is empty (cfg.DatabaseURL)")
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database using DSN '%s': %v", dsn, err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Database ping failed using DSN '%s': %v", dsn, err)
	}

	return db
}
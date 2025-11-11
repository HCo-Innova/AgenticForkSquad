package main

import (
	"context"
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
	internalcfg "github.com/tuusuario/afs-challenge/internal/config"
	repo "github.com/tuusuario/afs-challenge/internal/infrastructure/database/repositories"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
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
	
	// Validate Tiger Cloud only if enabled
	if cfg.UseTigerCloud {
		if err := cfg.ValidateTiger(); err != nil {
			log.Fatalf("Tiger Cloud config error: %v", err)
		}
		applogger.Info("✅ Tiger Cloud configuration validated")
	} else {
		applogger.Info("ℹ️  Tiger Cloud disabled - using direct database connections")
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
	userRepo := repo.NewPostgresUserRepository(db)

	// ============================================
	// 5. Initialize Use Cases (Application Layer)
	// ============================================
	// orchestrateMigrationUC := usecases.NewOrchestrateMigrationUseCase(migrationRepo)
	authService := usecases.NewAuthService(userRepo, cfg.JWTSecret)

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
	agentExecRepo := repo.NewPostgresAgentExecutionRepository(db)
	optRepo := repo.NewPostgresOptimizationRepository(db)
	benchRepo := repo.NewPostgresBenchmarkRepository(db)
	consRepo := repo.NewPostgresConsensusRepository(db)
	
	// Inicializar servicios y processors
	taskSvc := usecases.NewTaskService(taskRepo)
	orchestrator := usecases.NewOrchestrator()
	consensus := usecases.NewConsensusEngine()
	
	// Convertir config simple a internal/config
	internalConfig := &internalcfg.Config{}
	internalConfig.Database.URL = cfg.DatabaseURL
	internalConfig.TigerCloud.UseTigerCloud = cfg.UseTigerCloud
	internalConfig.TigerCloud.MainService = cfg.TigerMainService
	internalConfig.TigerCloud.MCPURL = cfg.TigerMCPURL
	internalConfig.VertexAI.ProjectID = os.Getenv("GCP_PROJECT_ID")
	internalConfig.VertexAI.Location = os.Getenv("GCP_REGION")
	internalConfig.VertexAI.Credentials = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	internalConfig.VertexAI.ModelCerebro = "gemini-2.5-pro"
	internalConfig.VertexAI.ModelOperativo = "gemini-2.5-flash"
	internalConfig.VertexAI.ModelBulk = "gemini-2.0-flash"
	
	// Inicializar MCP Client para Tiger Cloud
	mcpClient, err := mcp.New(internalConfig, nil)
	if err != nil {
		applogger.Info("⚠️ MCP Client initialization failed: " + err.Error())
		applogger.Info("Continuing without Tiger Cloud integration...")
		mcpClient = nil
	} else {
		applogger.Info("✅ MCP Client initialized successfully")
		// Conectar MCP (esto puede fallar si tiger CLI no está disponible)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mcpClient.Connect(ctx); err != nil {
			applogger.Info("⚠️ MCP Client connection failed: " + err.Error())
			applogger.Info("Continuing with limited MCP functionality...")
		} else {
			applogger.Info("✅ MCP Client connected successfully")
		}
	}
	
	// Inicializar Orchestrator con MCP Client
	orchestrator.MCPClient = mcpClient
	
	// Crear AgentFactory con MCP Client
	agentFactory := usecases.NewAgentFactory(mcpClient, agentExecRepo, internalConfig)
	
	// Crear TaskProcessor con todas las dependencias
	var taskProcessor *usecases.TaskProcessor
	if mcpClient != nil {
		taskProcessor = usecases.NewTaskProcessor(
			taskRepo,
			agentExecRepo,
			optRepo,
			benchRepo,
			consRepo,
			orchestrator,
			consensus,
			hub,
			agentFactory,
			internalConfig.TigerCloud.MainService,
		)
		applogger.Info("✅ TaskProcessor initialized with full agent processing")
	} else {
		applogger.Info("⚠️ TaskProcessor disabled (MCP not available)")
		taskProcessor = nil
	}
	
	// Handlers
	taskHandler := handlers.NewTaskHandler(taskSvc, taskProcessor, hub)
	resultsHandler := handlers.NewResultsHandler(agentExecRepo, optRepo, benchRepo, consRepo, hub)
	authHandler := handlers.NewAuthHandler(authService)
	metricsHandler := handlers.NewMetricsHandler(db)
	routes.SetupRoutes(app, hub, taskHandler, resultsHandler, authHandler, authService, metricsHandler)

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
	listenAddr := "0.0.0.0:" + cfg.Port
	applogger.Info("🚀 Server starting on " + listenAddr)
	applogger.Info("Environment: " + cfg.Environment)

	if err := app.Listen(listenAddr); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// ============================================
// Helper Functions
// ============================================

func getAllowedOrigins(env string) string {
	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins != "" {
		return origins
	}
	if env == "production" {
		return "*"
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
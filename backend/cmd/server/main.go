package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/tuusuario/afs-challenge/internal/config"
	llm "github.com/tuusuario/afs-challenge/internal/infrastructure/llm"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
	repo "github.com/tuusuario/afs-challenge/internal/infrastructure/database/repositories"
	httphandlers "github.com/tuusuario/afs-challenge/internal/presentation/http/handlers"
	"github.com/tuusuario/afs-challenge/internal/presentation/http/routes"
	"github.com/tuusuario/afs-challenge/internal/usecases"
)

func main() {
	// 1) Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// 2) Connect to database
	db, err := sqlx.Open("postgres", cfg.Database.URL)
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}
	defer db.Close()
	if err := pingDBWithTimeout(db, 10*time.Second); err != nil {
		log.Fatalf("db ping error: %v", err)
	}

	// 3) Run migrations (optional)
	if os.Getenv("RUN_MIGRATIONS") == "true" {
		log.Println("[migrations] RUN_MIGRATIONS=true → ejecutar migraciones (no-op placeholder)")
		// TODO: integrar runner real de migraciones si corresponde.
	}

	// 4) Initialize repositories
	taskRepo := repo.NewPostgresTaskRepository(db)
	agentExecRepo := repo.NewPostgresAgentExecutionRepository(db)
	_ = agentExecRepo // wired later where needed
	optRepo := repo.NewPostgresOptimizationRepository(db)
	_ = optRepo
	benchRepo := repo.NewPostgresBenchmarkRepository(db)
	_ = benchRepo
	consRepo := repo.NewPostgresConsensusRepository(db)
	_ = consRepo

	// 5) Initialize MCP client
	mcpClient, err := mcp.New(cfg, nil)
	if err != nil { log.Fatalf("mcp init error: %v", err) }
	if cfg.TigerCloud.UseTigerCloud {
		if err := mcpClient.Connect(context.Background()); err != nil {
			log.Fatalf("mcp connect error: %v", err)
		}
	}

	// 6) Initialize LLM clients (Vertex unified) por rol
	llmCerebro, err := llm.NewVertexClient(cfg, cfg.VertexAI.ModelCerebro, nil)
	if err != nil { log.Fatalf("vertex cerebro error: %v", err) }
	llmOperativo, err := llm.NewVertexClient(cfg, cfg.VertexAI.ModelOperativo, nil)
	if err != nil { log.Fatalf("vertex operativo error: %v", err) }
	llmBulk, err := llm.NewVertexClient(cfg, cfg.VertexAI.ModelBulk, nil)
	if err != nil { log.Fatalf("vertex bulk error: %v", err) }
	_ = llmCerebro
	_ = llmOperativo
	_ = llmBulk

	// 8) Initialize WebSocket Hub
	hub := usecases.NewHub()
	go hub.Run()

	// 9) Initialize Use Cases
	taskSvc := usecases.NewTaskService(taskRepo)
	agentFactory := usecases.NewAgentFactory(mcpClient, agentExecRepo, cfg)
	consEngine := usecases.NewConsensusEngine()
	orch := usecases.NewOrchestrator()
	taskProcessor := usecases.NewTaskProcessor(
		taskRepo,
		agentExecRepo,
		optRepo,
		benchRepo,
		consRepo,
		orch,
		consEngine,
		hub,
		agentFactory,
		cfg.TigerCloud.MainService,
	)

	// 10) Initialize HTTP Handlers and Router
	app := fiber.New()
	taskHandler := httphandlers.NewTaskHandler(taskSvc, taskProcessor, hub)
	resultsHandler := httphandlers.NewResultsHandler(agentExecRepo, optRepo, benchRepo, consRepo, hub)
	authHandler := httphandlers.NewAuthHandler(nil)
	metricsHandler := httphandlers.NewMetricsHandler(db)
	authSvc := usecases.NewAuthService(nil, "dummy-jwt-secret")
	
	// CORS middleware - allow Vercel frontend
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Set("Access-Control-Allow-Credentials", "true")
		if c.Method() == "OPTIONS" {
			return c.SendStatus(204)
		}
		return c.Next()
	})
	
	routes.SetupRoutes(app, hub, taskHandler, resultsHandler, authHandler, authSvc, metricsHandler)

	// 11) Start HTTP/WebSocket server
	addr := net.JoinHostPort(cfg.Server.Host, cfg.Server.Port)
	go func() {
		log.Printf("server listening on http://%s", addr)
		if err := app.Listen(addr); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	<-sigC
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = app.Shutdown()
	_ = db.Close()
	_ = ctx
}

func pingDBWithTimeout(db *sqlx.DB, d time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	return db.PingContext(ctx)
}

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	cfgpkg "github.com/tuusuario/afs-challenge/internal/config"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/agents"
	llm "github.com/tuusuario/afs-challenge/internal/infrastructure/llm"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
	repo "github.com/tuusuario/afs-challenge/internal/infrastructure/database/repositories"
	httphandlers "github.com/tuusuario/afs-challenge/internal/presentation/http/handlers"
	"github.com/tuusuario/afs-challenge/internal/presentation/http/routes"
	"github.com/tuusuario/afs-challenge/internal/usecases"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
)

// agentFactory implements usecases.AgentFactory backed by pre-wired dependencies.
type agentFactory struct {
	mcp  *mcp.MCPClient
	llms map[values.AgentType]llm.LLMClient
	cfg  *cfgpkg.Config
}

func (f *agentFactory) New(t values.AgentType) (agents.Agent, error) {
	client, ok := f.llms[t]
	if !ok || client == nil {
		return nil, fmt.Errorf("no LLM client for agent type %s", t)
	}
	return agents.NewAgent(t, f.mcp, client, f.cfg)
}

func main() {
	// 1) Load config
	cfg, err := cfgpkg.Load()
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

	// 7) Initialize Agent Implementations and Factory
	factory := &agentFactory{
		mcp: mcpClient,
		cfg: cfg,
		llms: map[values.AgentType]llm.LLMClient{
			values.AgentCerebro:  llmCerebro,
			values.AgentOperativo: llmOperativo,
			values.AgentBulk:      llmBulk,
		},
	}

	// 8) Initialize WebSocket Hub
	hub := usecases.NewHub()
	go hub.Run()

	// 9) Initialize Use Cases
	taskSvc := usecases.NewTaskService(taskRepo)
	router := usecases.NewRouter(factory)
	benchRunner := usecases.NewBenchmarkRunner(mcpClient)
	consEngine := usecases.NewConsensusEngine()
	orch := usecases.NewOrchestrator()
	// Wire Orchestrator MCP for later phases (apply/cleanup)
	// Note: Orchestrator fields are intentionally minimal; tests use injected methods
	_ = benchRunner
	_ = consEngine
	_ = router
	_ = taskSvc
	_ = hub
	_ = orch

	// 10) Initialize HTTP Handlers and Router
	app := fiber.New()
	taskHandler := httphandlers.NewTaskHandler(taskSvc, hub)
	resultsHandler := httphandlers.NewResultsHandler(agentExecRepo, optRepo, benchRepo, consRepo, hub)
	routes.SetupRoutes(app, hub, taskHandler, resultsHandler)

	// Middleware básicos (CORS/logging) pueden agregarse aquí si están implementados
	// app.Use(middleware.CORS())
	// app.Use(middleware.Logging())

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

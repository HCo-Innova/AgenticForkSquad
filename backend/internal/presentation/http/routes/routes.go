package routes

import (
    "github.com/gofiber/fiber/v2"
    websocket "github.com/gofiber/websocket/v2"
    "github.com/tuusuario/afs-challenge/internal/presentation/http/handlers"
    "github.com/tuusuario/afs-challenge/internal/presentation/http/middleware"
    "github.com/tuusuario/afs-challenge/internal/usecases"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, hub *usecases.Hub, taskH *handlers.TaskHandler, resH *handlers.ResultsHandler, authH *handlers.AuthHandler, authSvc *usecases.AuthService, metricsH *handlers.MetricsHandler) {
    // ============================================
    // Health Check Endpoints
    // ============================================
    app.Get("/health", handlers.HealthCheck)
    app.Get("/health/live", handlers.Liveness)
    app.Get("/health/ready", handlers.Readiness)

    // ============================================
    // API v1 Routes
    // ============================================
    api := app.Group("/api/v1")

    // ============================================
    // Authentication (Public - under /api/v1)
    // ============================================
    auth := api.Group("/auth")
    auth.Post("/login", authH.Login)
    auth.Post("/register", authH.Register)
    auth.Get("/me", middleware.AuthMiddleware(authSvc), authH.Me)

    // Root endpoint
    api.Get("/", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "message": "AFS Challenge API",
            "version": "1.0.0",
            "status":  "active",
        })
    })

    // ============================================
    // Metrics (Public)
    // ============================================
    api.Get("/metrics/overview", metricsH.GetOverview)
    api.Get("/metrics/agents", metricsH.GetAgentMetrics)
    api.Get("/metrics/performance", metricsH.GetPerformance)

    // ============================================
    // Tasks
    // ============================================
    api.Post("/tasks", taskH.CreateTask)
    api.Get("/tasks", taskH.ListTasks)
    api.Get("/tasks/:id", taskH.GetTask)
    api.Delete("/tasks/:id", taskH.DeleteTask)
    api.Get("/tasks/:id/agents", resH.GetTaskAgents)
    api.Get("/tasks/:id/proposals", resH.GetTaskProposals)
    api.Get("/tasks/:id/consensus", resH.GetTaskConsensus)

    // ============================================
    // Proposals
    // ============================================
    api.Get("/proposals/:id/benchmarks", resH.GetProposalBenchmarks)

    // ============================================
    // Agents
    // ============================================
    agents := api.Group("/agents")
    agents.Get("/", resH.ListAllAgents)
    agents.Get("/:type/status", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"message": "Agent status - TODO", "type": c.Params("type")})
    })

    // ============================================
    // WebSocket
    // ============================================
    app.Get("/ws", websocket.New(handlers.NewWSHandler(hub)))

    // ============================================
    // 404 Handler
    // ============================================
    app.Use(func(c *fiber.Ctx) error {
        return c.Status(404).JSON(fiber.Map{
            "error":   "Not Found",
            "message": "The requested resource does not exist",
            "path":    c.Path(),
        })
    })
}

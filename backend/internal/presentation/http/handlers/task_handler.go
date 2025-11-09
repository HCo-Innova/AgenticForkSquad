package handlers

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/usecases"
)

// TaskHandler encapsula los endpoints REST de tareas.
type TaskHandler struct {
	TaskService    *usecases.TaskService
	TaskProcessor  *usecases.TaskProcessor
	Hub            *usecases.Hub
}

func NewTaskHandler(svc *usecases.TaskService, processor *usecases.TaskProcessor, hub *usecases.Hub) *TaskHandler {
	return &TaskHandler{
		TaskService:   svc,
		TaskProcessor: processor,
		Hub:           hub,
	}
}

// ======================
// DTOs de solicitud
// ======================

type createTaskRequest struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	TargetQuery string                 `json:"target_query"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ======================
// Helpers de mapeo
// ======================

func mapEntityToResponse(t *entities.Task) fiber.Map {
	// Campos base según 08-API-SPECIFICATION.md
	resp := fiber.Map{
		"id":           t.ID,
		"type":         t.Type,
		"description":  t.Description,
		"target_query": t.TargetQuery,
		"status":       t.Status,
		"created_at":   t.CreatedAt.Format(time.RFC3339),
		"completed_at": nil,
		"metadata":     t.Metadata,
	}
	if t.CompletedAt != nil {
		resp["completed_at"] = t.CompletedAt.Format(time.RFC3339)
	}
	resp["links"] = fiber.Map{
		"self":      "/api/v1/tasks/" + strconv.FormatInt(t.ID, 10),
		"agents":    "/api/v1/tasks/" + strconv.FormatInt(t.ID, 10) + "/agents",
		"proposals": "/api/v1/tasks/" + strconv.FormatInt(t.ID, 10) + "/proposals",
	}
	return resp
}

// ======================
// Handlers
// ======================

// POST /api/v1/tasks
func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	if h == nil || h.TaskService == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":      "INTERNAL_ERROR",
				"message":   "Task service not available",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
	}

	var req createTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":      "VALIDATION_ERROR",
				"message":   "Invalid request body",
				"details":   fiber.Map{"body": "cannot parse JSON"},
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
	}

	// Validaciones mínimas según spec
	if strings.TrimSpace(req.Type) == "" || strings.TrimSpace(req.TargetQuery) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Missing required fields",
				"details": fiber.Map{
					"type":         "required",
					"target_query": "required",
				},
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
	}

	ent := &entities.Task{
		Type:        entities.TaskType(req.Type),
		Description: req.Description,
		TargetQuery: req.TargetQuery,
		Status:      entities.TaskStatusPending,
		CreatedAt:   time.Now().UTC(),
		Metadata:    req.Metadata,
	}

	created, err := h.TaskService.CreateTask(c.Context(), ent)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":      "VALIDATION_ERROR",
				"message":   err.Error(),
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
	}

	// Emitir evento WS: task_created
	if h.Hub != nil {
		h.Hub.Broadcast(usecases.Event{
			Type: usecases.EventTaskCreated,
			Payload: map[string]interface{}{
				"id":           created.ID,
				"type":         created.Type,
				"status":       created.Status,
				"created_at":   created.CreatedAt.Format(time.RFC3339),
				"target_query": created.TargetQuery,
			},
		})
	}

	// Procesar tarea asíncronamente
	if h.TaskProcessor != nil {
		go func(taskID int64) {
			ctx := context.Background()
			if err := h.TaskProcessor.ProcessTask(ctx, taskID); err != nil {
				// Log error pero no fallar el request HTTP
				// El error se reflejará en el estado de la tarea
				println("Error processing task", taskID, ":", err.Error())
			}
		}(int64(created.ID))
	}

	return c.Status(fiber.StatusCreated).JSON(mapEntityToResponse(created))
}

// GET /api/v1/tasks/:id
func (h *TaskHandler) GetTask(c *fiber.Ctx) error {
	if h == nil || h.TaskService == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":      "INTERNAL_ERROR",
				"message":   "Task service not available",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
	}
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":      "VALIDATION_ERROR",
				"message":   "Invalid id parameter",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
	}
	t, err := h.TaskService.GetTask(c.Context(), id)
	if err != nil || t == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":      "TASK_NOT_FOUND",
				"message":   "Task not found",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
	}
	return c.JSON(mapEntityToResponse(t))
}

// GET /api/v1/tasks
func (h *TaskHandler) ListTasks(c *fiber.Ctx) error {
	if h == nil || h.TaskService == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":      "INTERNAL_ERROR",
				"message":   "Task service not available",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
	}

	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	offset := 0
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	filters := entities.TaskFilters{
		Status:        c.Query("status"),
		Type:          c.Query("type"),
		CreatedAfter:  c.Query("created_after"),
		CreatedBefore: c.Query("created_before"),
	}

	list, err := h.TaskService.ListTasks(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":      "INTERNAL_ERROR",
				"message":   err.Error(),
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
	}

	// Paginación simple en memoria (el repository puede manejarla mejor).
	start := offset
	if start > len(list) {
		start = len(list)
	}
	end := start + limit
	if end > len(list) {
		end = len(list)
	}
	paged := list[start:end]

	data := make([]fiber.Map, 0, len(paged))
	for _, t := range paged {
		data = append(data, mapEntityToResponse(t))
	}

	return c.JSON(fiber.Map{
		"data": data,
		"pagination": fiber.Map{
			"limit":    limit,
			"offset":   offset,
			"total":    len(list),
			"has_more": end < len(list),
		},
	})
}

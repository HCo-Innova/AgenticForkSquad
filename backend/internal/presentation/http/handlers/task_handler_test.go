package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	domainif "github.com/tuusuario/afs-challenge/internal/domain/interfaces"
	"github.com/tuusuario/afs-challenge/internal/usecases"
)

type fakeTaskRepo struct {
	stored []*entities.Task
}

func TestGetTask_InvalidID(t *testing.T) {
    app := fiber.New()
    h := newTaskHandlerWithFake()
    app.Get("/api/v1/tasks/:id", h.GetTask)

    req := httptest.NewRequest("GET", "/api/v1/tasks/abc", nil)
    resp, _ := app.Test(req)
    if resp.StatusCode != 400 { t.Fatalf("expected 400, got %d", resp.StatusCode) }
}

func TestCreateTask_ServiceNil(t *testing.T) {
    app := fiber.New()
    h := &TaskHandler{TaskService: nil}
    app.Post("/api/v1/tasks", h.CreateTask)

    req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader([]byte(`{}`)))
    req.Header.Set("Content-Type", "application/json")
    resp, _ := app.Test(req)
    if resp.StatusCode != 500 { t.Fatalf("expected 500, got %d", resp.StatusCode) }
}

func TestListTasks_RepoError(t *testing.T) {
    app := fiber.New()
    r := &errRepo{}
    svc := usecases.NewTaskService(r)
    h := NewTaskHandler(svc, nil)
    app.Get("/api/v1/tasks", h.ListTasks)

    req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
    resp, _ := app.Test(req)
    if resp.StatusCode != 500 { t.Fatalf("expected 500, got %d", resp.StatusCode) }
}

func TestCreateTask_InvalidJSON(t *testing.T) {
    app := fiber.New()
    h := newTaskHandlerWithFake()
    app.Post("/api/v1/tasks", h.CreateTask)

    req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader([]byte("{invalid")))
    req.Header.Set("Content-Type", "application/json")
    resp, _ := app.Test(req)
    if resp.StatusCode != 400 { t.Fatalf("expected 400, got %d", resp.StatusCode) }
}

func (f *fakeTaskRepo) Create(ctx context.Context, t *entities.Task) error {
	t.ID = 1
	f.stored = append(f.stored, t)
	return nil
}
func (f *fakeTaskRepo) GetByID(ctx context.Context, id int) (*entities.Task, error) {
	for _, t := range f.stored {
		if int(t.ID) == id { return t, nil }
	}
	return nil, nil
}
func (f *fakeTaskRepo) List(ctx context.Context, filters entities.TaskFilters) ([]*entities.Task, error) {
	return f.stored, nil
}
func (f *fakeTaskRepo) Update(ctx context.Context, t *entities.Task) error { return nil }

// Ensure fake implements interface at compile time
var _ domainif.TaskRepository = (*fakeTaskRepo)(nil)

// errRepo simula un error en List para probar la ruta 500 de ListTasks
type errRepo struct{ fakeTaskRepo }
func (e *errRepo) List(ctx context.Context, filters entities.TaskFilters) ([]*entities.Task, error) {
    return nil, errors.New("repo error")
}

func newTaskHandlerWithFake() *TaskHandler {
	repo := &fakeTaskRepo{}
	svc := usecases.NewTaskService(repo)
	return NewTaskHandler(svc, nil)
}

func TestCreateTask_HappyPath(t *testing.T) {
	app := fiber.New()
	h := newTaskHandlerWithFake()
	app.Post("/api/v1/tasks", h.CreateTask)

	body := map[string]any{
		"type": "query_optimization",
		"description": "desc",
		"target_query": "SELECT 1",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil { t.Fatalf("request failed: %v", err) }
	if resp.StatusCode != 201 { t.Fatalf("expected 201, got %d", resp.StatusCode) }
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { t.Fatalf("invalid json: %v", err) }
	if out["id"] == nil { t.Fatalf("expected id in response") }
}

func TestCreateTask_ValidationError(t *testing.T) {
	app := fiber.New()
	h := newTaskHandlerWithFake()
	app.Post("/api/v1/tasks", h.CreateTask)

	body := map[string]any{ "description": "desc" }
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != 400 { t.Fatalf("expected 400, got %d", resp.StatusCode) }
}

func TestGetTask_NotFound(t *testing.T) {
	app := fiber.New()
	h := newTaskHandlerWithFake()
	app.Get("/api/v1/tasks/:id", h.GetTask)

	req := httptest.NewRequest("GET", "/api/v1/tasks/123", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 404 { t.Fatalf("expected 404, got %d", resp.StatusCode) }
}

func TestListTasks_PaginationBasic(t *testing.T) {
	app := fiber.New()
	h := newTaskHandlerWithFake()
	// seed two tasks
	h.TaskService.CreateTask(nil, &entities.Task{ID: 1, Type: "query_optimization", TargetQuery: "SELECT 1", Status: entities.TaskStatusPending, CreatedAt: time.Now()})
	h.TaskService.CreateTask(nil, &entities.Task{ID: 2, Type: "query_optimization", TargetQuery: "SELECT 2", Status: entities.TaskStatusPending, CreatedAt: time.Now()})
	app.Get("/api/v1/tasks", h.ListTasks)

	req := httptest.NewRequest("GET", "/api/v1/tasks?limit=1&offset=0", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 { t.Fatalf("expected 200, got %d", resp.StatusCode) }
}

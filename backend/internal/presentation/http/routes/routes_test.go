package routes

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/tuusuario/afs-challenge/internal/presentation/http/handlers"
)

func TestRoutes_RootAnd404(t *testing.T) {
	app := fiber.New()
	// minimal handlers for wiring
	SetupRoutes(app, nil, &handlers.TaskHandler{}, &handlers.ResultsHandler{})

	req := httptest.NewRequest("GET", "/api/v1/", nil)
	resp, err := app.Test(req)
	if err != nil { t.Fatalf("request failed: %v", err) }
	if resp.StatusCode != 200 { t.Fatalf("expected 200, got %d", resp.StatusCode) }
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil { t.Fatalf("invalid json: %v", err) }
	if body["message"] != "AFS Challenge API" { t.Fatalf("unexpected body: %+v", body) }

	req = httptest.NewRequest("GET", "/does-not-exist", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != 404 { t.Fatalf("expected 404, got %d", resp.StatusCode) }
}

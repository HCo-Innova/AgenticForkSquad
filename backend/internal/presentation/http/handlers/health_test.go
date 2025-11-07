package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestHealthEndpoints(t *testing.T) {
	app := fiber.New()
	app.Get("/health", HealthCheck)
	app.Get("/health/live", Liveness)
	app.Get("/health/ready", Readiness)

	// /health
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("/health request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["status"] == "" {
		t.Errorf("missing status in response")
	}

	// /health/live
	req = httptest.NewRequest("GET", "/health/live", nil)
	resp, err = app.Test(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("/health/live failed: %v code=%d", err, resp.StatusCode)
	}

	// /health/ready (helpers return nil -> ready)
	req = httptest.NewRequest("GET", "/health/ready", nil)
	resp, err = app.Test(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("/health/ready failed: %v code=%d", err, resp.StatusCode)
	}
}

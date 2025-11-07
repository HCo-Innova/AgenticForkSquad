package config

import (
	"os"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	os.Setenv("PORT", "8000")
	os.Setenv("ENV", "development")
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Server.Port != "8000" || cfg.Server.Environment != "development" {
		t.Error("server config not loaded correctly")
	}
}

func TestLoadBuildsDatabaseURLFromPostgresEnv(t *testing.T) {
	t.Setenv("PORT", "8000")
	t.Setenv("ENV", "development")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("POSTGRES_DB", "afs_dev")
	t.Setenv("POSTGRES_USER", "afs_user")
	t.Setenv("POSTGRES_PASSWORD", "afs_pass")
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_PORT", "5432")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Database.URL == "" {
		t.Error("expected database URL to be built from POSTGRES_* envs")
	}
}

func TestTigerCloudValidation(t *testing.T) {
	t.Setenv("PORT", "8000")
	t.Setenv("ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")

	// Enable Tiger Cloud but omit required fields
	t.Setenv("USE_TIGER_CLOUD", "true")
	t.Setenv("TIGER_MAIN_SERVICE", "")
	t.Setenv("TIGER_MCP_URL", "")

	if _, err := Load(); err == nil {
		t.Error("expected validation error when Tiger Cloud enabled without required fields")
	}
}

func TestLoadMissingVars(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("ENV")

	_, err := Load()
	if err == nil {
		t.Error("expected error for missing required variables")
	}
}

func TestVertexModelDefaultsWhenUnset(t *testing.T) {
	t.Setenv("PORT", "8000")
	t.Setenv("ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")

	// Unset role-specific envs to trigger defaults
	t.Setenv("GEMINI_CEREBRO_MODEL", "")
	t.Setenv("GEMINI_OPERATIVO_MODEL", "")
	t.Setenv("GEMINI_BULK_MODEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.VertexAI.ModelCerebro != "gemini-2.5-pro" {
		t.Errorf("expected default cerebromodel, got %s", cfg.VertexAI.ModelCerebro)
	}
	if cfg.VertexAI.ModelOperativo != "gemini-2.5-flash" {
		t.Errorf("expected default operativo model, got %s", cfg.VertexAI.ModelOperativo)
	}
	if cfg.VertexAI.ModelBulk != "gemini-2.0-flash" {
		t.Errorf("expected default bulk model, got %s", cfg.VertexAI.ModelBulk)
	}
}

func TestDBMaxConnectionsParsing(t *testing.T) {
	t.Setenv("PORT", "8000")
	t.Setenv("ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")

	// Valid positive integer
	t.Setenv("DB_MAX_CONNECTIONS", "25")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Database.MaxConnections != 25 {
		t.Errorf("expected 25, got %d", cfg.Database.MaxConnections)
	}

	// Invalid value → ignore and fallback to default 10
	t.Setenv("DB_MAX_CONNECTIONS", "not-a-number")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Database.MaxConnections != 10 {
		t.Errorf("expected default 10 for invalid value, got %d", cfg.Database.MaxConnections)
	}

	// Zero or negative → fallback to default 10
	t.Setenv("DB_MAX_CONNECTIONS", "0")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Database.MaxConnections != 10 {
		t.Errorf("expected default 10 for zero, got %d", cfg.Database.MaxConnections)
	}
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTigerConfig_FileNotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	_, err := LoadTigerConfig()
	if err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
}

func TestLoadTigerConfig_InvalidJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	confDir := filepath.Join(home, ".config", "tiger")
	if err := os.MkdirAll(confDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	confPath := filepath.Join(confDir, "mcp-config.json")
	if err := os.WriteFile(confPath, []byte("{invalid-json"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
	_, err := LoadTigerConfig()
	if err == nil {
		t.Fatalf("expected JSON unmarshal error, got nil")
	}
}

func TestLoadTigerConfig_Valid(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	confDir := filepath.Join(home, ".config", "tiger")
	if err := os.MkdirAll(confDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	confPath := filepath.Join(confDir, "mcp-config.json")
	data := `{
		"main_service": "afs-main",
		"mcp_url": "https://mcp.tiger",
		"auth_token": "secret"
	}`
	if err := os.WriteFile(confPath, []byte(data), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
	cfg, err := LoadTigerConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MainService != "afs-main" || cfg.MCPURL != "https://mcp.tiger" || cfg.AuthToken != "secret" {
		t.Fatalf("parsed config mismatch: %+v", cfg)
	}
}

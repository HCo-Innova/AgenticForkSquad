package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type TigerConfig struct {
	MainService string `json:"main_service"`
	MCPURL      string `json:"mcp_url"`
	AuthToken   string `json:"auth_token"`
}

// LoadTigerConfig reads ~/.config/tiger/mcp-config.json if present.
func LoadTigerConfig() (*TigerConfig, error) {
	path := filepath.Join(os.Getenv("HOME"), ".config", "tiger", "mcp-config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg TigerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

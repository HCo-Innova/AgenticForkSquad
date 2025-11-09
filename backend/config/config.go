package config

import (
	"errors"
	"os"
)

type Config struct {
	Environment string
	Port        string
	DatabaseURL string
	RedisURL    string
	LogLevel    string
	UseTigerCloud   bool
	TigerMainService string
	TigerMCPURL      string
	JWTSecret       string
}

func buildTigerURLFromEnv() string {
    host := os.Getenv("TIGER_DB_HOST")
    port := os.Getenv("TIGER_DB_PORT")
    user := os.Getenv("TIGER_DB_USER")
    pass := os.Getenv("TIGER_DB_PASSWORD")
    db := os.Getenv("TIGER_DB_NAME")
    if host == "" || port == "" || user == "" || pass == "" || db == "" {
        return ""
    }
    return "postgresql://" + user + ":" + pass + "@" + host + ":" + port + "/" + db + "?sslmode=require"
}

func Load() *Config {
    dsn := buildTigerURLFromEnv()
    if dsn == "" {
        dsn = getEnv("DATABASE_URL", "")
    }
    return &Config{
        Environment:     getEnv("ENV", "development"),
        Port:            getEnv("PORT", "8000"),
        DatabaseURL:     dsn,
        RedisURL:        getEnv("REDIS_URL", ""),
        LogLevel:        getEnv("LOG_LEVEL", "info"),
        UseTigerCloud:   getEnvBool("USE_TIGER_CLOUD", false),
        TigerMainService: getEnv("TIGER_MAIN_SERVICE", ""),
        TigerMCPURL:      getEnv("TIGER_MCP_URL", ""),
        JWTSecret:        getEnv("JWT_SECRET", "default-secret-change-in-production"),
    }
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	switch v {
	case "1", "true", "TRUE", "True", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "False", "no", "NO", "off", "OFF":
		return false
	default:
		return defaultVal
	}
}

func (c *Config) ValidateTiger() error {
	if !c.UseTigerCloud {
		return nil
	}
	if c.TigerMainService == "" {
		return errors.New("TIGER_MAIN_SERVICE is required when USE_TIGER_CLOUD=true")
	}
	if c.TigerMCPURL == "" {
		return errors.New("TIGER_MCP_URL is required when USE_TIGER_CLOUD=true")
	}
	return nil
}

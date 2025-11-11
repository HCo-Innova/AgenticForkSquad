package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server struct {
		Host        string
		Port        string
		Environment string
		LogLevel    string
	}
	Database struct {
		URL              string
		MaxConnections   int
		MaxIdleConns     int
		ConnMaxLifetimeS int
		ConnMaxIdleTimeS int
	}
	Redis struct {
		URL      string
		Password string
	}
	VertexAI struct {
		ProjectID      string
		Location       string
		ModelCerebro   string    // gemini-2.5-pro (Planner/QA)
		ModelOperativo string    // gemini-2.5-flash (Generación/Ejecución)
		ModelBulk      string    // gemini-2.0-flash (Bajo costo)
		Credentials    string
	}
	TigerCloud struct {
		UseTigerCloud bool
		MainService   string
		ForkAgent1    string
		ForkAgent2    string
		MCPURL        string
		PublicKey     string
		SecretKey     string
		ProjectID     string
	}
	Timeouts struct {
		LLMAnalysisMS  int
		LLMProposalMS  int
		MCPQueryMS     int
		DBConnectMS    int
		ContextMS      int
	}
}

// Load reads configuration from environment variables and validates required fields.
func Load() (*Config, error) {
	cfg := &Config{}

	// Server
	cfg.Server.Host = os.Getenv("HOST")
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0" // Default to bind all interfaces for containers
	}
	cfg.Server.Port = os.Getenv("PORT")
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8000" // Default for Railway
	}
	cfg.Server.Environment = os.Getenv("ENV")
	cfg.Server.LogLevel = os.Getenv("LOG_LEVEL")

	// Database
	// Prefer TIGER_DB_* (cloud) if present or explicitly requested
	if tigerURL := buildTigerURLFromEnv(); tigerURL != "" {
		cfg.Database.URL = tigerURL
	} else {
		cfg.Database.URL = os.Getenv("DATABASE_URL")
		if cfg.Database.URL == "" {
			cfg.Database.URL = buildDatabaseURLFromEnv()
		}
	}
	if v := os.Getenv("DB_MAX_CONNECTIONS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Database.MaxConnections = n
		}
	}
	if cfg.Database.MaxConnections == 0 {
		cfg.Database.MaxConnections = 10
	}
	if v := os.Getenv("DB_MAX_IDLE_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Database.MaxIdleConns = n
		}
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 5
	}
	if v := os.Getenv("DB_CONN_MAX_LIFETIME_S"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Database.ConnMaxLifetimeS = n
		}
	}
	if cfg.Database.ConnMaxLifetimeS == 0 {
		cfg.Database.ConnMaxLifetimeS = 3600
	}
	if v := os.Getenv("DB_CONN_MAX_IDLE_TIME_S"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Database.ConnMaxIdleTimeS = n
		}
	}
	if cfg.Database.ConnMaxIdleTimeS == 0 {
		cfg.Database.ConnMaxIdleTimeS = 600
	}

	// Redis
	cfg.Redis.URL = os.Getenv("REDIS_URL")
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")

	// Vertex AI (Solo Gemini models)
	cfg.VertexAI.ProjectID = os.Getenv("VERTEX_PROJECT_ID")
	cfg.VertexAI.Location = os.Getenv("VERTEX_LOCATION")
	// Modelos por rol (con defaults seguros a Gemini)
	if v := os.Getenv("GEMINI_CEREBRO_MODEL"); v != "" {
		cfg.VertexAI.ModelCerebro = v
	} else {
		cfg.VertexAI.ModelCerebro = "gemini-2.5-pro"
	}
	if v := os.Getenv("GEMINI_OPERATIVO_MODEL"); v != "" {
		cfg.VertexAI.ModelOperativo = v
	} else {
		cfg.VertexAI.ModelOperativo = "gemini-2.5-flash"
	}
	if v := os.Getenv("GEMINI_BULK_MODEL"); v != "" {
		cfg.VertexAI.ModelBulk = v
	} else {
		cfg.VertexAI.ModelBulk = "gemini-2.0-flash"
	}
	cfg.VertexAI.Credentials = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	// Tiger Cloud
	cfg.TigerCloud.UseTigerCloud = os.Getenv("USE_TIGER_CLOUD") == "true"
	cfg.TigerCloud.MainService = os.Getenv("TIGER_MAIN_SERVICE")
	// Fork IDs (try new naming first, fallback to old naming)
	if forkA1 := os.Getenv("TIGER_FORK_A1_SERVICE_ID"); forkA1 != "" {
		cfg.TigerCloud.ForkAgent1 = forkA1
	} else {
		cfg.TigerCloud.ForkAgent1 = os.Getenv("TIGER_FORK_AGENT_1")
	}
	if forkA2 := os.Getenv("TIGER_FORK_A2_SERVICE_ID"); forkA2 != "" {
		cfg.TigerCloud.ForkAgent2 = forkA2
	} else {
		cfg.TigerCloud.ForkAgent2 = os.Getenv("TIGER_FORK_AGENT_2")
	}
	cfg.TigerCloud.MCPURL = os.Getenv("TIGER_MCP_URL")
	cfg.TigerCloud.PublicKey = os.Getenv("TIGER_PUBLIC_KEY")
	cfg.TigerCloud.SecretKey = os.Getenv("TIGER_SECRET_KEY")
	cfg.TigerCloud.ProjectID = os.Getenv("TIGER_PROJECT_ID")

	// Timeouts (milliseconds)
	if v := os.Getenv("TIMEOUT_LLM_ANALYSIS_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Timeouts.LLMAnalysisMS = n
		}
	}
	if cfg.Timeouts.LLMAnalysisMS == 0 {
		cfg.Timeouts.LLMAnalysisMS = 120000 // 120s
	}
	if v := os.Getenv("TIMEOUT_LLM_PROPOSAL_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Timeouts.LLMProposalMS = n
		}
	}
	if cfg.Timeouts.LLMProposalMS == 0 {
		cfg.Timeouts.LLMProposalMS = 60000 // 60s
	}
	if v := os.Getenv("TIMEOUT_MCP_QUERY_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Timeouts.MCPQueryMS = n
		}
	}
	if cfg.Timeouts.MCPQueryMS == 0 {
		cfg.Timeouts.MCPQueryMS = 60000 // 60s
	}
	if v := os.Getenv("TIMEOUT_DB_CONNECT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Timeouts.DBConnectMS = n
		}
	}
	if cfg.Timeouts.DBConnectMS == 0 {
		cfg.Timeouts.DBConnectMS = 10000 // 10s
	}
	if v := os.Getenv("TIMEOUT_CONTEXT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Timeouts.ContextMS = n
		}
	}
	if cfg.Timeouts.ContextMS == 0 {
		cfg.Timeouts.ContextMS = 600000 // 10m
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.Server.Environment == "" {
		return errors.New("missing required environment variable: ENV")
	}
	if c.Database.URL == "" {
		return errors.New("missing required database configuration: set DATABASE_URL or POSTGRES_DB/USER/PASSWORD")
	}
	if c.TigerCloud.UseTigerCloud {
		if strings.TrimSpace(c.TigerCloud.MainService) == "" {
			return errors.New("missing TIGER_MAIN_SERVICE when USE_TIGER_CLOUD=true")
		}
		if strings.TrimSpace(c.TigerCloud.ForkAgent1) == "" {
			return errors.New("missing TIGER_FORK_AGENT_1 when USE_TIGER_CLOUD=true")
		}
		if strings.TrimSpace(c.TigerCloud.ForkAgent2) == "" {
			return errors.New("missing TIGER_FORK_AGENT_2 when USE_TIGER_CLOUD=true")
		}
	}
	return nil
}

func buildDatabaseURLFromEnv() string {
	db := os.Getenv("POSTGRES_DB")
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "postgres"
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	if db == "" || user == "" || pass == "" {
		return ""
	}
	return "postgres://" + user + ":" + pass + "@" + host + ":" + port + "/" + db
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

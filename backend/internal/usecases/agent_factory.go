package usecases

import (
	"fmt"

	cfgpkg "github.com/tuusuario/afs-challenge/internal/config"
	domainif "github.com/tuusuario/afs-challenge/internal/domain/interfaces"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/agents"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/llm"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
)

// AgentFactory crea instancias de agentes con las dependencias necesarias
type AgentFactory struct {
	MCPClient    *mcp.MCPClient
	AgentRepo    domainif.AgentExecutionRepository
	Cfg          *cfgpkg.Config
}

// NewAgentFactory creates a new agent factory with required dependencies
func NewAgentFactory(mcpClient *mcp.MCPClient, agentRepo domainif.AgentExecutionRepository, cfg *cfgpkg.Config) *AgentFactory {
	return &AgentFactory{
		MCPClient: mcpClient,
		AgentRepo: agentRepo,
		Cfg:       cfg,
	}
}

// CreateAgent crea un agente del tipo especificado con su cliente LLM correspondiente
func (f *AgentFactory) CreateAgent(agentType values.AgentType) (agents.Agent, error) {
	if f.MCPClient == nil || f.Cfg == nil || f.AgentRepo == nil {
		return nil, fmt.Errorf("agent factory not properly initialized")
	}

	// Determinar modelo según tipo de agente
	var model string
	switch agentType {
	case values.AgentCerebro:
		model = f.Cfg.VertexAI.ModelCerebro // gemini-2.5-pro
		if model == "" {
			model = "gemini-2.5-pro"
		}
	case values.AgentOperativo:
		model = f.Cfg.VertexAI.ModelOperativo // gemini-2.5-flash
		if model == "" {
			model = "gemini-2.5-flash"
		}
	case values.AgentBulk:
		model = f.Cfg.VertexAI.ModelBulk // gemini-2.0-flash
		if model == "" {
			model = "gemini-2.0-flash"
		}
	default:
		return nil, fmt.Errorf("unknown agent type: %s", agentType)
	}

	// Crear cliente LLM para este agente
	llmClient, err := llm.NewVertexClient(f.Cfg, model, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client for %s: %w", model, err)
	}

	// Crear agente usando la factory del paquete agents
	agent, err := agents.NewAgent(agentType, f.MCPClient, llmClient, f.Cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return agent, nil
}

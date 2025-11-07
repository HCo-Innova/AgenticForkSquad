package usecases

import (
	"context"
	"errors"
	"strings"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/agents"
)

// AgentFactory abstracts agent creation for routing.
type AgentFactory interface {
	New(agentType values.AgentType) (agents.Agent, error)
}

// Router selects appropriate agents for a given task based on rules.
type Router struct {
	Factory   AgentFactory
	Rationale string
}

func NewRouter(factory AgentFactory) *Router { return &Router{Factory: factory} }

// SelectAgents applies routing rules and returns agent instances.
func (r *Router) SelectAgents(ctx context.Context, task *entities.Task) ([]agents.Agent, error) {
	if r == nil || r.Factory == nil {
		return nil, errors.New("router not initialized")
	}
	if task == nil {
		return nil, errors.New("task is required")
	}

	chosen := map[values.AgentType]bool{}
	reasons := []string{}

	priority := ""
	if task.Metadata != nil {
		if v, ok := task.Metadata["priority"].(string); ok {
			priority = strings.ToLower(strings.TrimSpace(v))
		}
	}
	if priority == "high" {
		chosen[values.AgentCerebro] = true
		chosen[values.AgentOperativo] = true
		chosen[values.AgentBulk] = true
		reasons = append(reasons, "high priority → cerebro+operativo+bulk")
	}

	q := strings.ToLower(task.TargetQuery)
	if strings.Contains(q, " join ") || strings.Contains(q, " join\n") {
		chosen[values.AgentOperativo] = true
		chosen[values.AgentCerebro] = true
		reasons = append(reasons, "JOIN detected → include operativo + cerebro")
	}

	var rows float64
	if task.Metadata != nil {
		switch v := task.Metadata["table_rows"].(type) {
		case float64:
			rows = v
		case int:
			rows = float64(v)
		}
		if rows == 0 {
			if v, ok := task.Metadata["table_size_rows"].(float64); ok {
				rows = v
			}
		}
	}
	if rows > 1_000_000 {
		chosen[values.AgentOperativo] = true
		reasons = append(reasons, ">1M rows → include operativo")
	}

	// Default at least operativo if nothing else chosen
	if len(chosen) == 0 {
		chosen[values.AgentOperativo] = true
		reasons = append(reasons, "default → operativo")
	}

	aList := make([]agents.Agent, 0, len(chosen))
	for t := range chosen {
		a, err := r.Factory.New(t)
		if err != nil { return nil, err }
		aList = append(aList, a)
	}
	r.Rationale = strings.Join(reasons, "; ")
	return aList, nil
}

package agents

import (
	"testing"

	cfgpkg "github.com/tuusuario/afs-challenge/internal/config"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
)

type dummyLLM struct{}
func (d *dummyLLM) SendMessage(prompt, system string) (string, error) { return "", nil }
func (d *dummyLLM) SendMessageWithJSON(prompt, system string) (map[string]interface{}, error) { return map[string]interface{}{}, nil }
func (d *dummyLLM) GetUsage() (int, int) { return 0, 0 }

func TestFactory(t *testing.T) {
	cfg := &cfgpkg.Config{}
	mcpClient, _ := mcp.New(cfg, nil)
	llmClient := &dummyLLM{}

	cases := []values.AgentType{
		values.AgentCerebro,
		values.AgentOperativo,
		values.AgentBulk,
	}
	for _, at := range cases {
		ag, err := NewAgent(at, mcpClient, llmClient, cfg)
		if err != nil { t.Fatalf("unexpected err for %s: %v", at, err) }
		if ag == nil { t.Fatalf("nil agent for %s", at) }
	}
	// invalid
	if _, err := NewAgent(values.AgentType("invalid"), mcpClient, llmClient, cfg); err == nil {
		t.Fatalf("expected error for invalid agent type")
	}
}

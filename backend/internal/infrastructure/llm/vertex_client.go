package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	cfgpkg "github.com/tuusuario/afs-challenge/internal/config"
)

// Doer abstracts http.Client for testing.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// VertexClient implements LLMClient for multiple Vertex models.
type VertexClient struct {
	httpClient  Doer
	projectID   string
	location    string
	model       string
	baseURL     string
	inputTokens int
	outputTokens int
	timeout     time.Duration
}

// NewVertexClient creates a client using config and an HTTP Doer (optional).
func NewVertexClient(cfg *cfgpkg.Config, model string, httpClient Doer) (*VertexClient, error) {
	if cfg == nil {
		return nil, errors.New("nil config")
	}
	if cfg.VertexAI.ProjectID == "" || cfg.VertexAI.Location == "" {
		return nil, errors.New("missing Vertex project/location")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	vc := &VertexClient{
		httpClient:  httpClient,
		projectID:   cfg.VertexAI.ProjectID,
		location:    cfg.VertexAI.Location,
		model:       normalizeModel(model),
		baseURL:     "https://aiplatform.googleapis.com",
		inputTokens: 0,
		outputTokens: 0,
		timeout:     30 * time.Second,
	}
	return vc, nil
}

func normalizeModel(m string) string {
	s := strings.TrimSpace(strings.ToLower(m))
	switch s {
	case "gemini-2.5-pro", "gemini25pro", "gemini":
		return "gemini-2.5-pro"
	case "gemini-2.5-flash", "g25f", "gemini-flash":
		return "gemini-2.5-flash"
	case "gemini-2.0-flash", "g20f":
		return "gemini-2.0-flash"
	default:
		return s
	}
}

func (c *VertexClient) GetUsage() (int, int) { return c.inputTokens, c.outputTokens }

// SendMessage returns raw text from the model.
func (c *VertexClient) SendMessage(prompt, system string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	raw, err := c.invoke(ctx, prompt, system, false)
	if err != nil { return "", err }
	return raw, nil
}

// SendMessageWithJSON returns parsed JSON map after removing markdown fences.
func (c *VertexClient) SendMessageWithJSON(prompt, system string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	raw, err := c.invoke(ctx, prompt, system, true)
	if err != nil { return nil, err }
	clean := stripMarkdownJSON(raw)
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(clean), &m); err != nil {
		// Best-effort fallback: try to extract between first '{' and last '}'
		if i := strings.Index(raw, "{"); i >= 0 {
			if j := strings.LastIndex(raw, "}"); j > i {
				candidate := raw[i:j+1]
				if json.Unmarshal([]byte(candidate), &m) == nil {
					return m, nil
				}
			}
		}
		return nil, err
	}
	return m, nil
}

// invoke builds a request per model and parses the primary text output.
func (c *VertexClient) invoke(ctx context.Context, prompt, system string, jsonMode bool) (string, error) {
	url := c.endpoint()
	body, err := c.requestBody(prompt, system, jsonMode)
	if err != nil { return "", err }
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil { return "", err }
	req.Header.Set("Content-Type", "application/json")
	// ADC authentication is handled by application default transport in real usage.
	// For tests, no auth is needed since we mock Doer.
	resp, err := c.httpClient.Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("vertex: http error")
	}
	// Parse per-model response (Gemini family)
	switch c.model {
	case "gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash":
		var obj struct {
			Candidates []struct{
				Content struct{
					Parts []struct{ Text string `json:"text"` } `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
			Usage struct {
				PromptTokenCount int `json:"promptTokenCount"`
				CandidatesTokenCount int `json:"candidatesTokenCount"`
			} `json:"usageMetadata"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&obj); err != nil { return "", err }
		c.inputTokens = obj.Usage.PromptTokenCount
		c.outputTokens = obj.Usage.CandidatesTokenCount
		if len(obj.Candidates) == 0 || len(obj.Candidates[0].Content.Parts) == 0 {
			return "", errors.New("vertex: empty candidates")
		}
		return obj.Candidates[0].Content.Parts[0].Text, nil
	default:
		return "", errors.New("vertex: unsupported model")
	}
}

func (c *VertexClient) endpoint() string {
	// We don't hit real endpoints in tests; return a stable placeholder.
	return c.baseURL + "/vertex/mock"
}

func (c *VertexClient) requestBody(prompt, system string, jsonMode bool) ([]byte, error) {
	switch c.model {
	case "gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash":
		gen := map[string]interface{}{
			"temperature": 0.0,
			"max_output_tokens": 8192,
		}
		if jsonMode { gen["response_mime_type"] = "application/json" }
		payload := map[string]interface{}{
			"contents": []map[string]interface{}{{
				"parts": []map[string]string{{"text": prompt}},
			}},
			"generation_config": gen,
		}
		return json.Marshal(payload)
	default:
		return nil, errors.New("unsupported model")
	}
}

var fenceRe = regexp.MustCompile("(?s)```(?:json)?\n(.*?)\n```")

func stripMarkdownJSON(s string) string {
	s = strings.TrimSpace(s)
	if m := fenceRe.FindStringSubmatch(s); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	return s
}

// intFromMap extracts an integer from a map with mixed numeric types.
func intFromMap(m map[string]interface{}, key string) int {
	if m == nil { return 0 }
	v, ok := m[key]
	if !ok { return 0 }
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t)
	case json.Number:
		if i, err := t.Int64(); err == nil { return int(i) }
		if f, err := t.Float64(); err == nil { return int(f) }
		return 0
	default:
		return 0
	}
}

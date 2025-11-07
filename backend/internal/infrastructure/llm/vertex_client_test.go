package llm

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	cfgpkg "github.com/tuusuario/afs-challenge/internal/config"
)

type mockDoer struct{
	responses []*http.Response
	errs      []error
	calls     int
}

func (m *mockDoer) Do(req *http.Request) (*http.Response, error) {
	idx := m.calls
	m.calls++
	if idx < len(m.errs) && m.errs[idx] != nil {
		return nil, m.errs[idx]
	}
	if idx < len(m.responses) {
		return m.responses[idx], nil
	}
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(`{"text":"{}"}`)))}, nil
}

func httpResp(status int, body interface{}) *http.Response {
	b, _ := json.Marshal(body)
	return &http.Response{StatusCode: status, Body: ioutil.NopCloser(bytes.NewReader(b))}
}

func TestVertexClient(t *testing.T) {
	cfg := &cfgpkg.Config{}
	cfg.VertexAI.ProjectID = "proj"
	cfg.VertexAI.Location = "us-central1"

	// Gemini 2.5 Pro: candidates plain JSON string
	geminiBodyPro := map[string]interface{}{
		"candidates": []map[string]interface{}{
			{"content": map[string]interface{}{
				"parts": []map[string]interface{}{{"text": "{\n  \"engine\": \"gemini\"\n}"}},
			}},
		},
		"usageMetadata": map[string]interface{}{"promptTokenCount": 30, "candidatesTokenCount": 11},
	}
	mdGeminiPro := &mockDoer{responses: []*http.Response{httpResp(200, geminiBodyPro)}}
	gm, err := NewVertexClient(cfg, "gemini-2.5-pro", mdGeminiPro)
	if err != nil { t.Fatalf("new gemini: %v", err) }
	jsonMap2, err := gm.SendMessageWithJSON("please json", "sys")
	if err != nil { t.Fatalf("gemini json parse err: %v", err) }
	if jsonMap2["engine"] != "gemini" { t.Fatalf("unexpected json: %v", jsonMap2) }

	// Gemini 2.5 Flash: candidates with fenced JSON
	geminiBodyFlash := map[string]interface{}{
		"candidates": []map[string]interface{}{
			{"content": map[string]interface{}{
				"parts": []map[string]interface{}{{"text": "```json\n{\n  \"tier\": \"flash\"\n}\n```"}},
			}},
		},
		"usageMetadata": map[string]interface{}{"promptTokenCount": 20, "candidatesTokenCount": 7},
	}
	mdGeminiFlash := &mockDoer{responses: []*http.Response{httpResp(200, geminiBodyFlash)}}
	gf, err := NewVertexClient(cfg, "gemini-2.5-flash", mdGeminiFlash)
	if err != nil { t.Fatalf("new gemini flash: %v", err) }
	jm, err := gf.SendMessageWithJSON("please json", "sys")
	if err != nil || jm["tier"] != "flash" { t.Fatalf("gemini flash json err: %v %v", err, jm) }

	// Gemini 2.0 Flash: candidates plain JSON string
	geminiBody20 := map[string]interface{}{
		"candidates": []map[string]interface{}{
			{"content": map[string]interface{}{
				"parts": []map[string]interface{}{{"text": "{\n  \"tier\": \"20\"\n}"}},
			}},
		},
		"usageMetadata": map[string]interface{}{"promptTokenCount": 12, "candidatesTokenCount": 3},
	}
	mdGemini20 := &mockDoer{responses: []*http.Response{httpResp(200, geminiBody20)}}
	g20, err := NewVertexClient(cfg, "gemini-2.0-flash", mdGemini20)
	if err != nil { t.Fatalf("new gemini 2.0: %v", err) }
	jm20, err := g20.SendMessageWithJSON("please json", "sys")
	if err != nil || jm20["tier"] != "20" { t.Fatalf("gemini 2.0 json err: %v %v", err, jm20) }
}

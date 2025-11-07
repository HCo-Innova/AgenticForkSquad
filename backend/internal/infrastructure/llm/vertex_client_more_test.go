package llm

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	cfgpkg "github.com/tuusuario/afs-challenge/internal/config"
)

type mockDoer2 struct{ resp *http.Response; err error }

func TestVertexClient_MultiPartCandidates_AndMissingUsage(t *testing.T) {
    cfg := &cfgpkg.Config{ VertexAI: struct{ ProjectID string; Location string; ModelCerebro string; ModelOperativo string; ModelBulk string; Credentials string }{ ProjectID: "p", Location: "l" } }
    // Response with two parts: should take the first part text
    payload := map[string]interface{}{
        "candidates": []map[string]interface{}{{
            "content": map[string]interface{}{
                "parts": []map[string]interface{}{
                    {"text": "{\n  \"pick\": 1\n}"},
                    {"text": "{\n  \"pick\": 2\n}"},
                },
            },
        }},
        // usageMetadata intentionally omitted to assert zeros
    }
    md := &mockDoer2{resp: &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(mustJSON(payload)))}}
    vc, _ := NewVertexClient(cfg, "gemini-2.5-pro", md)
    m, err := vc.SendMessageWithJSON("x", "y")
    if err != nil { t.Fatalf("unexpected err: %v", err) }
    if m["pick"] != float64(1) { t.Fatalf("expected first part pick=1, got %v", m["pick"]) }
    in, out := vc.GetUsage(); if in != 0 || out != 0 { t.Fatalf("expected zero usage, got %d %d", in, out) }
}
func (m *mockDoer2) Do(req *http.Request) (*http.Response, error) { return m.resp, m.err }

func TestNormalizeModelAliases(t *testing.T) {
	cases := map[string]string{
		"gemini": "gemini-2.5-pro",
		"gemini25pro": "gemini-2.5-pro",
		"g25f": "gemini-2.5-flash",
		"gemini-flash": "gemini-2.5-flash",
		"g20f": "gemini-2.0-flash",
	}
	for in, want := range cases {
		if got := normalizeModel(in); got != want {
			t.Fatalf("normalize %s => %s (got %s)", in, want, got)
		}
	}
}

func TestVertexClient_HTTPErrorAndEmptyCandidates(t *testing.T) {
	cfg := &cfgpkg.Config{}
	cfg.VertexAI.ProjectID = "p"; cfg.VertexAI.Location = "l"

	// HTTP error branch
	mdErr := &mockDoer2{resp: &http.Response{StatusCode: 500, Body: ioutil.NopCloser(bytes.NewReader([]byte(`{}`)))}}
	vc, _ := NewVertexClient(cfg, "gemini-2.5-pro", mdErr)
	if _, err := vc.SendMessage("x", "y"); err == nil { t.Fatalf("expected http error") }

	// Empty candidates branch
	body := map[string]interface{}{"candidates": []interface{}{}, "usageMetadata": map[string]interface{}{"promptTokenCount":1,"candidatesTokenCount":1}}
	mdEmpty := &mockDoer2{resp: &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(mustJSON(body)))}}
	vc2, _ := NewVertexClient(cfg, "gemini-2.5-pro", mdEmpty)
	if _, err := vc2.SendMessage("x", "y"); err == nil { t.Fatalf("expected empty candidates error") }
}

func TestVertexClient_JSONFallbackAndIntFromMap(t *testing.T) {
	cfg := &cfgpkg.Config{}
	cfg.VertexAI.ProjectID = "p"; cfg.VertexAI.Location = "l"
	// malformed JSON but contains {...}
	payload := map[string]interface{}{
		"candidates": []map[string]interface{}{{
			"content": map[string]interface{}{"parts": []map[string]interface{}{{"text": "prefix {\"ok\": true} suffix"}}},
		}},
		"usageMetadata": map[string]interface{}{"promptTokenCount": json.Number("12"), "candidatesTokenCount": json.Number("34")},
	}
	md := &mockDoer2{resp: &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(mustJSON(payload)))}}
	vc, _ := NewVertexClient(cfg, "gemini-2.5-pro", md)
	m, err := vc.SendMessageWithJSON("x", "y")
	if err != nil || m["ok"] != true { t.Fatalf("fallback json failed %v %v", err, m) }
	in, out := vc.GetUsage(); if in != 12 || out != 34 { t.Fatalf("usage mismatch %d %d", in, out) }
}

func TestVertexClient_UnsupportedModel(t *testing.T) {
	cfg := &cfgpkg.Config{ VertexAI: struct{ ProjectID string; Location string; ModelCerebro string; ModelOperativo string; ModelBulk string; Credentials string }{ ProjectID: "p", Location: "l" } }
	md := &mockDoer2{resp: &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(`{}`)))}}
	vc, _ := NewVertexClient(cfg, "unknown-model", md)
	if _, err := vc.SendMessage("x", "y"); err == nil { t.Fatalf("expected unsupported model error") }
}

func mustJSON(v interface{}) []byte { b, _ := json.Marshal(v); return b }

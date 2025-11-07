package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	cfgpkg "github.com/tuusuario/afs-challenge/internal/config"
)

type mockDoer struct{
	responses []*http.Response
	errs      []error
	calls     int
}

func TestMCPClient_MissingBaseURL(t *testing.T) {
    cfg := &cfgpkg.Config{}
    md := &mockDoer{}
    client, _ := New(cfg, md)
    ctx := context.Background()
    if err := client.Connect(ctx); err == nil {
        t.Fatalf("expected error due to missing base URL")
    }
}

func TestMCPClient_HTTPErrorWithMessage(t *testing.T) {
    md := &mockDoer{responses: []*http.Response{
        newErrorResp(toolResponse{Success:false, Error:&toolError{Message:"boom"}}, 500),
        newErrorResp(toolResponse{Success:false, Error:&toolError{Message:"boom"}}, 500),
        newErrorResp(toolResponse{Success:false, Error:&toolError{Message:"boom"}}, 500),
    }}
    cfg := &cfgpkg.Config{}
    cfg.TigerCloud.MCPURL = "http://mock.local"
    client, _ := New(cfg, md)
    ctx := context.Background()
    if err := client.Connect(ctx); err == nil || err.Error() != "mcp: request failed after retries" {
        t.Fatalf("expected final retry error, got %v", err)
    }
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
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(`{"success":true}`)))}, nil
}

func newSuccessResp(body interface{}) *http.Response {
	b, _ := json.Marshal(body)
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(b))}
}

func newErrorResp(body interface{}, status int) *http.Response {
	b, _ := json.Marshal(body)
	return &http.Response{StatusCode: status, Body: ioutil.NopCloser(bytes.NewReader(b))}
}

func TestMCPClient_Connect_OK(t *testing.T) {
	md := &mockDoer{
		responses: []*http.Response{
			newSuccessResp(toolResponse{Success: true, Data: map[string]interface{}{"services": []interface{}{}, "total": 0}}),
		},
	}
	cfg := &cfgpkg.Config{}
	cfg.TigerCloud.MCPURL = "http://mock.local"
	client, err := New(cfg, md)
	if err != nil { t.Fatalf("new: %v", err) }
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("connect err: %v", err)
	}
}

func TestMCPClient_ExecuteQuery_OK(t *testing.T) {
	payload := toolResponse{Success: true, Data: map[string]interface{}{
		"rows": []map[string]interface{}{{"email":"a@b.com","revenue":"10.0"}},
		"row_count": 1,
		"execution_time_ms": 12.3,
	}}
	md := &mockDoer{responses: []*http.Response{newSuccessResp(payload)}}
	cfg := &cfgpkg.Config{}
	cfg.TigerCloud.MCPURL = "http://mock.local"
	client, _ := New(cfg, md)
	ctx := context.Background()
	res, err := client.ExecuteQuery(ctx, "svc-1", "SELECT 1", 1000)
	if err != nil { t.Fatalf("execute err: %v", err) }
	if res.RowCount != 1 || len(res.Rows) != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestMCPClient_RetryOnTransportError(t *testing.T) {
	md := &mockDoer{
		errs: []error{context.DeadlineExceeded, context.DeadlineExceeded},
		responses: []*http.Response{newSuccessResp(toolResponse{Success:true})},
	}
	cfg := &cfgpkg.Config{}
	cfg.TigerCloud.MCPURL = "http://mock.local"
	client, _ := New(cfg, md)
	client.maxRetries = 3
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if md.calls < 3 {
		t.Fatalf("expected retries, calls=%d", md.calls)
	}
}

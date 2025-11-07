package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	cfgpkg "github.com/tuusuario/afs-challenge/internal/config"
)

type mockDoer2 struct{
	responses []*http.Response
	errs      []error
	calls     int
}

func (m *mockDoer2) Do(req *http.Request) (*http.Response, error) {
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

func httpResp(status int, body interface{}) *http.Response {
	b, _ := json.Marshal(body)
	return &http.Response{StatusCode: status, Body: ioutil.NopCloser(bytes.NewReader(b))}
}

func TestFork_CreateDelete_List_Info(t *testing.T) {
	// 1) CreateFork -> success
	createBody := toolResponse{Success: true, Data: map[string]interface{}{"service_id": "afs-fork-test"}}
	// 2) ListForks -> one service
	listBody := toolResponse{Success: true, Data: map[string]interface{}{
		"services": []map[string]interface{}{{"service_id":"afs-fork-test","status":"active"}},
	}}
	// 3) GetServiceInfo -> success (we ignore rich payload)
	infoBody := toolResponse{Success: true, Data: map[string]interface{}{"schema":"public"}}
	// 4) DeleteFork -> success
	deleteBody := toolResponse{Success: true}

	md := &mockDoer2{responses: []*http.Response{
		httpResp(200, createBody),
		httpResp(200, listBody),
		httpResp(200, infoBody),
		httpResp(200, deleteBody),
	}}
	cfg := &cfgpkg.Config{}
	cfg.TigerCloud.MCPURL = "http://mock.local"
	client, _ := New(cfg, md)

	ctx := context.Background()
	id, err := client.CreateFork(ctx, "afs-main", "afs-fork-test")
	if err != nil || id == "" { t.Fatalf("CreateFork err=%v id=%s", err, id) }

	services, err := client.ListForks(ctx, "afs-main", 0, 0)
	if err != nil || len(services) != 1 { t.Fatalf("ListForks err=%v n=%d", err, len(services)) }
	if services[0].ServiceID != "afs-fork-test" { t.Fatalf("unexpected service id: %s", services[0].ServiceID) }

	info, err := client.GetServiceInfo(ctx, id)
	if err != nil || info.ServiceID != id { t.Fatalf("GetServiceInfo err=%v info=%+v", err, info) }

	if err := client.DeleteFork(ctx, id); err != nil { t.Fatalf("DeleteFork err=%v", err) }
}

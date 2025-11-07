package usecases

import (
	"encoding/json"
	"testing"
	"time"
)

func TestWebSocketHub(t *testing.T) {
	h := NewHub()
	go h.Run()

	c1 := h.Register()
	c2 := h.Register()

	h.Broadcast(Event{Type: "task_started", Payload: map[string]interface{}{"id": 1}})

	select {
	case msg := <-c1:
		var evt Event
		_ = json.Unmarshal(msg, &evt)
		if evt.Type != "task_started" { t.Fatalf("unexpected event: %s", evt.Type) }
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting for broadcast on client 1")
	}

	select {
	case msg := <-c2:
		var evt Event
		_ = json.Unmarshal(msg, &evt)
		if evt.Type != "task_started" { t.Fatalf("unexpected event: %s", evt.Type) }
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting for broadcast on client 2")
	}

	h.Unregister(c1)
	h.Unregister(c2)
}

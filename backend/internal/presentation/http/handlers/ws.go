package handlers

import (
	"encoding/json"
	"time"
	websocket "github.com/gofiber/websocket/v2"
	"github.com/tuusuario/afs-challenge/internal/usecases"
)

// NewWSHandler returns a Fiber WebSocket handler bound to the shared Hub.
func NewWSHandler(hub *usecases.Hub) func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		client := hub.Register()
		defer hub.Unregister(client)

		// Per-connection subscription filter (empty means receive all)
		subscribed := map[string]bool{}

		// Push a welcome message via the hub
		hub.Broadcast(usecases.Event{
			Type:    "connection_established",
			Payload: map[string]interface{}{"message": "Connected to AFS WebSocket"},
		})

		// Forward hub broadcasts to this websocket connection
		done := make(chan struct{})
		go func() {
			for msg := range client {
				// Filter by subscription if configured
				if len(subscribed) > 0 {
					var ev usecases.Event
					if err := json.Unmarshal(msg, &ev); err == nil {
						if !subscribed[ev.Type] {
							continue
						}
					}
				}
				_ = c.WriteMessage(websocket.TextMessage, msg)
			}
			close(done)
		}()

		// Read loop: support optional client->server messages: ping, subscribe
		for {
			mt, raw, err := c.ReadMessage()
			if err != nil {
				break
			}
			if mt != websocket.TextMessage {
				continue
			}

			var msg struct {
				Type    string                 `json:"type"`
				Payload map[string]interface{} `json:"payload"`
			}
			if err := json.Unmarshal(raw, &msg); err != nil {
				continue
			}

			switch msg.Type {
			case "ping":
				// Respond directly to the client with a pong
				pong, _ := json.Marshal(usecases.Event{
					Type: "pong",
					Payload: map[string]interface{}{
						"timestamp": time.Now().UTC().Format(time.RFC3339),
					},
				})
				_ = c.WriteMessage(websocket.TextMessage, pong)
			case "subscribe":
				// Expect payload { "events": ["event_a", ...] }
				if evs, ok := msg.Payload["events"].([]interface{}); ok {
					// reset and set
					subscribed = map[string]bool{}
					for _, v := range evs {
						if s, ok := v.(string); ok && s != "" {
							subscribed[s] = true
						}
					}
					ack, _ := json.Marshal(usecases.Event{
						Type: "subscribed",
						Payload: map[string]interface{}{
							"events": evs,
						},
					})
					_ = c.WriteMessage(websocket.TextMessage, ack)
				}
			default:
				// Unknown types are ignored for now
			}
		}

		<-done
	}
}

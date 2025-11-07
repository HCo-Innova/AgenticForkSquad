package usecases

import (
	"encoding/json"
)

// Event represents a WS event (simplified per doc 08)
type Event struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type clientReg struct{
	c   chan []byte
	ack chan struct{}
}

// Hub maintains active clients and broadcasts events to all of them.
type Hub struct {
	clients    map[chan []byte]bool
	broadcast  chan []byte
	register   chan clientReg
	unregister chan chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[chan []byte]bool),
		broadcast:  make(chan []byte, 32),
		register:   make(chan clientReg, 32),
		unregister: make(chan chan []byte, 32),
	}
}

// Run processes client registration and broadcasts. Call as a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case reg := <-h.register:
			h.clients[reg.c] = true
			// signal caller that client is registered
			if reg.ack != nil { close(reg.ack) }
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c)
			}
		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c <- msg:
				default:
					// slow client → drop and unregister
					delete(h.clients, c)
					close(c)
				}
			}
		}
	}
}

// Broadcast encodes the event to JSON and enqueues it.
func (h *Hub) Broadcast(e Event) {
	b, _ := json.Marshal(e)
	h.broadcast <- b
}

// Register returns a client channel registered in the hub.
func (h *Hub) Register() chan []byte {
	c := make(chan []byte, 8)
	ack := make(chan struct{})
	h.register <- clientReg{c: c, ack: ack}
	// block until the hub confirms registration to avoid race in tests
	<-ack
	return c
}

// Unregister removes the client from the hub.
func (h *Hub) Unregister(c chan []byte) { h.unregister <- c }

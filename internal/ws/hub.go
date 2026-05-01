package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

const (
	EventNewPendingUser = "new_pending_user"
	EventUserApproved   = "user_approved"
	EventUserRejected   = "user_rejected"
)

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type client struct {
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
}

var Global = &Hub{
	clients: make(map[*client]struct{}),
}

func (h *Hub) register(c *client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) unregister(c *client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	close(c.send)
}

func (h *Hub) Broadcast(eventType string, payload interface{}) {
	data, err := json.Marshal(Event{Type: eventType, Payload: payload})
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients {
		select {
		case c.send <- data:
		default:
		}
	}
}

func ServeWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("ws upgrade error:", err)
		return
	}

	cl := &client{conn: conn, send: make(chan []byte, 64)}
	Global.register(cl)

	go func() {
		defer func() {
			conn.Close()
		}()
		for msg := range cl.send {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				break
			}
		}
	}()

	defer Global.unregister(cl)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

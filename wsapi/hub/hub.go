package hub

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn   *websocket.Conn
	UserID int32
	Send   chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[int32]*Client // key: userID
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int32]*Client),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.UserID] = client
}

func (h *Hub) Unregister(userID int32) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if client, ok := h.clients[userID]; ok {
		close(client.Send)
		delete(h.clients, userID)
	}
}

func (h *Hub) Get(userID int32) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, ok := h.clients[userID]
	return client, ok
}

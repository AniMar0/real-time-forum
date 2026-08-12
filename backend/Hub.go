package backend

import "sync"

// Hub owns the set of active WebSocket clients. Server code interacts with
// connections through this type instead of keeping connection state itself.
type Hub struct {
	mu      sync.RWMutex
	clients map[int64]map[string]*Client
}

func NewHub() *Hub {
	return &Hub{clients: make(map[int64]map[string]*Client)}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[string]*Client)
	}
	h.clients[client.UserID][client.ID] = client
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	if sessions := h.clients[client.UserID]; sessions != nil {
		delete(sessions, client.ID)
		if len(sessions) == 0 {
			delete(h.clients, client.UserID)
		}
	}
	h.mu.Unlock()

	client.Close()
}

func (h *Hub) ClientsForUser(userID int64) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	sessions := h.clients[userID]
	clients := make([]*Client, 0, len(sessions))
	for _, client := range sessions {
		clients = append(clients, client)
	}
	return clients
}

func (h *Hub) UserIDs() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userIDs := make([]int64, 0, len(h.clients))
	for userID := range h.clients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

func (h *Hub) IsOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID]) > 0
}

func (h *Hub) SendToUser(userID int64, message interface{}) {
	for _, client := range h.ClientsForUser(userID) {
		client.Enqueue(message)
	}
}

func (h *Hub) Broadcast(message interface{}) {
	for _, userID := range h.UserIDs() {
		h.SendToUser(userID, message)
	}
}

func (h *Hub) DisconnectSession(userID int64, sessionID string) {
	for _, client := range h.ClientsForUser(userID) {
		if client.SessionID == sessionID {
			h.Unregister(client)
		}
	}
}

func (c *Client) Enqueue(message interface{}) bool {
	c.sendMu.RLock()
	defer c.sendMu.RUnlock()

	select {
	case c.Send <- message:
		return true
	default:
		return false
	}
}

func (c *Client) Close() {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()

	c.closeOnce.Do(func() {
		close(c.Send)
		if c.Conn != nil {
			_ = c.Conn.Close()
		}
	})
}

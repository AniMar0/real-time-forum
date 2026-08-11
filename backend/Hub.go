package backend

import "sync"

// Hub owns the set of active WebSocket clients. Server code interacts with
// connections through this type instead of keeping connection state itself.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[string]*Client
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[string]*Client)}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.Username] == nil {
		h.clients[client.Username] = make(map[string]*Client)
	}
	h.clients[client.Username][client.ID] = client
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	if sessions := h.clients[client.Username]; sessions != nil {
		delete(sessions, client.ID)
		if len(sessions) == 0 {
			delete(h.clients, client.Username)
		}
	}
	h.mu.Unlock()

	client.Close()
}

func (h *Hub) ClientsForUser(username string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	sessions := h.clients[username]
	clients := make([]*Client, 0, len(sessions))
	for _, client := range sessions {
		clients = append(clients, client)
	}
	return clients
}

func (h *Hub) Usernames() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	usernames := make([]string, 0, len(h.clients))
	for username := range h.clients {
		usernames = append(usernames, username)
	}
	return usernames
}

func (h *Hub) IsOnline(username string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[username]) > 0
}

func (h *Hub) SendToUser(username string, message interface{}) {
	for _, client := range h.ClientsForUser(username) {
		client.Enqueue(message)
	}
}

func (h *Hub) Broadcast(message interface{}) {
	for _, username := range h.Usernames() {
		h.SendToUser(username, message)
	}
}

func (h *Hub) DisconnectSession(username, sessionID string) {
	for _, client := range h.ClientsForUser(username) {
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
		_ = c.Conn.Close()
	})
}

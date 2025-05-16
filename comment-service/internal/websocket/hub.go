package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	PostID   string
	UserID   int64
	closeMux sync.Mutex
	closed   bool
}

func NewClient(hub *Hub, conn *websocket.Conn, postID string, userID int64) *Client {
	return &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		PostID:   postID,
		UserID:   userID,
		closeMux: sync.Mutex{},
		closed:   false,
	}
}

type Hub struct {
	// Registered clients for each post
	clients map[string]map[*Client]bool

	// Channel for broadcasting messages to clients
	broadcast chan []byte

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if _, ok := h.clients[client.PostID]; !ok {
				h.clients[client.PostID] = make(map[*Client]bool)
			}
			h.clients[client.PostID][client] = true
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.PostID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.PostID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			var event struct {
				PostID string `json:"post_id"`
			}
			if err := json.Unmarshal(message, &event); err != nil {
				log.Printf("Error unmarshaling broadcast message: %v", err)
				continue
			}

			h.mu.RLock()
			clients := h.clients[event.PostID]
			h.mu.RUnlock()

			for client := range clients {
				select {
				case client.Send <- message:
				default:
					h.mu.Lock()
					delete(h.clients[client.PostID], client)
					close(client.Send)
					if len(h.clients[client.PostID]) == 0 {
						delete(h.clients, client.PostID)
					}
					h.mu.Unlock()
				}
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.closeMux.Lock()
		if !c.closed {
			c.Conn.Close()
			c.closed = true
		}
		c.closeMux.Unlock()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.closeMux.Lock()
		if !c.closed {
			c.Conn.Close()
			c.closed = true
		}
		c.closeMux.Unlock()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (h *Hub) BroadcastToPost(postID string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	h.broadcast <- data
	return nil
}

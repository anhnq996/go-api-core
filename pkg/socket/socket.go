package socket

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`
	Data      interface{}            `json:"data"`
	Room      string                 `json:"room,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Client represents a WebSocket client
type Client struct {
	ID     string
	UserID string
	Conn   *websocket.Conn
	Send   chan Message
	Rooms  map[string]bool
	Hub    *Hub
	mu     sync.RWMutex
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Rooms for grouping clients
	rooms map[string]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to all clients
	broadcast chan Message

	// Broadcast messages to specific room
	roomBroadcast chan Message

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]map[*Client]bool),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan Message),
		roomBroadcast: make(chan Message),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToAll(message)

		case message := <-h.roomBroadcast:
			h.broadcastToRoom(message)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
	log.Printf("Client %s connected. Total clients: %d", client.ID, len(h.clients))
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)

		// Remove client from all rooms
		for room := range client.Rooms {
			h.removeClientFromRoom(client, room)
		}

		log.Printf("Client %s disconnected. Total clients: %d", client.ID, len(h.clients))
	}
}

// broadcastToAll broadcasts message to all clients
func (h *Hub) broadcastToAll(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.clients, client)
		}
	}
}

// broadcastToRoom broadcasts message to specific room
func (h *Hub) broadcastToRoom(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if room, exists := h.rooms[message.Room]; exists {
		for client := range room {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.clients, client)
				delete(room, client)
			}
		}
	}
}

// JoinRoom adds client to a room
func (h *Hub) JoinRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][client] = true
	client.Rooms[room] = true

	log.Printf("Client %s joined room %s", client.ID, room)
}

// LeaveRoom removes client from a room
func (h *Hub) LeaveRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.removeClientFromRoom(client, room)
}

// removeClientFromRoom removes client from room (internal use)
func (h *Hub) removeClientFromRoom(client *Client, room string) {
	if roomClients, exists := h.rooms[room]; exists {
		delete(roomClients, client)
		delete(client.Rooms, room)

		// Clean up empty room
		if len(roomClients) == 0 {
			delete(h.rooms, room)
		}

		log.Printf("Client %s left room %s", client.ID, room)
	}
}

// GetRoomClients returns clients in a room
func (h *Hub) GetRoomClients(room string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	if roomClients, exists := h.rooms[room]; exists {
		for client := range roomClients {
			clients = append(clients, client)
		}
	}
	return clients
}

// GetClientCount returns total number of clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetRoomCount returns number of rooms
func (h *Hub) GetRoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}

// BroadcastToAll sends message to all clients
func (h *Hub) BroadcastToAll(message Message) {
	h.broadcast <- message
}

// BroadcastToRoom sends message to specific room
func (h *Hub) BroadcastToRoom(room string, message Message) {
	message.Room = room
	h.roomBroadcast <- message
}

// BroadcastToUser sends message to specific user
func (h *Hub) BroadcastToUser(userID string, message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	message.UserID = userID
	for client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		var message Message
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle different message types
		switch message.Type {
		case "join_room":
			if room, ok := message.Data.(string); ok {
				c.Hub.JoinRoom(c, room)
			}
		case "leave_room":
			if room, ok := message.Data.(string); ok {
				c.Hub.LeaveRoom(c, room)
			}
		case "broadcast":
			c.Hub.BroadcastToAll(message)
		case "room_message":
			if room, ok := message.Data.(map[string]interface{})["room"].(string); ok {
				c.Hub.BroadcastToRoom(room, message)
			}
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

// ServeWS handles websocket requests from clients
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow connections from any origin (configure as needed)
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Extract user ID from request (implement your authentication logic)
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	client := &Client{
		ID:     fmt.Sprintf("client_%d", len(hub.clients)),
		UserID: userID,
		Conn:   conn,
		Send:   make(chan Message, 256),
		Rooms:  make(map[string]bool),
		Hub:    hub,
	}

	client.Hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// MessageType defines the type of WebSocket message
type MessageType string

const (
	MessageTypeDeploymentStatus MessageType = "deployment_status"
	MessageTypeBuildLog         MessageType = "build_log"
	MessageTypeDeploymentStart  MessageType = "deployment_start"
	MessageTypeDeploymentEnd    MessageType = "deployment_end"
)

// Message represents a WebSocket message
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

// DeploymentStatusPayload contains deployment status update data
type DeploymentStatusPayload struct {
	DeploymentID uint   `json:"deploymentId"`
	ProjectID    uint   `json:"projectId"`
	Status       string `json:"status"`
	Error        string `json:"error,omitempty"`
}

// BuildLogPayload contains build log data
type BuildLogPayload struct {
	DeploymentID uint   `json:"deploymentId"`
	ProjectID    uint   `json:"projectId"`
	Message      string `json:"message"`
	Level        string `json:"level"`
	Timestamp    string `json:"timestamp"`
}

// Client represents a WebSocket client
type Client struct {
	Conn      *websocket.Conn
	UserID    uint
	ProjectID *uint // nil means subscribed to all projects
	Send      chan []byte
	hub       *Hub
}

// Hub maintains active WebSocket connections and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client connected (UserID: %d, ProjectID: %v). Total clients: %d",
				client.UserID, client.ProjectID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				log.Printf("WebSocket client disconnected (UserID: %d). Total clients: %d",
					client.UserID, len(h.clients))
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling WebSocket message: %v", err)
				h.mu.RUnlock()
				continue
			}

			// Broadcast to all matching clients
			for client := range h.clients {
				// Check if client should receive this message
				if h.shouldReceiveMessage(client, message) {
					select {
					case client.Send <- data:
					default:
						// Client's send buffer is full, disconnect them
						close(client.Send)
						delete(h.clients, client)
						log.Printf("WebSocket client send buffer full, disconnecting (UserID: %d)", client.UserID)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// shouldReceiveMessage determines if a client should receive a specific message
func (h *Hub) shouldReceiveMessage(client *Client, message *Message) bool {
	// If client is subscribed to all projects
	if client.ProjectID == nil {
		return true
	}

	// Check if message is project-specific
	switch message.Type {
	case MessageTypeDeploymentStatus:
		if payload, ok := message.Payload.(DeploymentStatusPayload); ok {
			return payload.ProjectID == *client.ProjectID
		}
	case MessageTypeBuildLog:
		if payload, ok := message.Payload.(BuildLogPayload); ok {
			return payload.ProjectID == *client.ProjectID
		}
	}

	return true
}

// BroadcastDeploymentStatus broadcasts a deployment status update
func (h *Hub) BroadcastDeploymentStatus(deploymentID, projectID uint, status string, errorMsg string) {
	h.broadcast <- &Message{
		Type: MessageTypeDeploymentStatus,
		Payload: DeploymentStatusPayload{
			DeploymentID: deploymentID,
			ProjectID:    projectID,
			Status:       status,
			Error:        errorMsg,
		},
	}
}

// BroadcastBuildLog broadcasts a build log message
func (h *Hub) BroadcastBuildLog(deploymentID, projectID uint, message, level, timestamp string) {
	h.broadcast <- &Message{
		Type: MessageTypeBuildLog,
		Payload: BuildLogPayload{
			DeploymentID: deploymentID,
			ProjectID:    projectID,
			Message:      message,
			Level:        level,
			Timestamp:    timestamp,
		},
	}
}

// ReadPump reads messages from the WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		// Currently, we don't expect messages from clients
		// In the future, we could handle client commands here
	}
}

// WritePump writes messages to the WebSocket connection
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Error writing WebSocket message: %v", err)
			return
		}
	}
}

// HandleWebSocket handles WebSocket upgrade and client management
func (h *Hub) HandleWebSocket(c *fiber.Ctx) error {
	// Get userID from context (set by auth middleware)
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}

	// Get optional projectID from query parameter
	var projectID *uint
	if projectIDStr := c.Query("projectId"); projectIDStr != "" {
		var pid uint
		if _, err := fiber.Scan(projectIDStr, &pid); err == nil {
			projectID = &pid
		}
	}

	// Check if this is a WebSocket upgrade request
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.NewError(fiber.StatusUpgradeRequired, "WebSocket upgrade required")
	}

	return websocket.New(func(conn *websocket.Conn) {
		// Create new client
		client := &Client{
			Conn:      conn,
			UserID:    userID,
			ProjectID: projectID,
			Send:      make(chan []byte, 256),
			hub:       h,
		}

		// Register client
		h.register <- client

		// Start read and write pumps
		go client.WritePump()
		client.ReadPump() // Blocks until connection is closed
	})(c)
}

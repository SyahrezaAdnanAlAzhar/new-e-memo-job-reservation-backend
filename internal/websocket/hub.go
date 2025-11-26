package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"e-memo-job-reservation-api/internal/repository"

	"github.com/gin-gonic/gin"
)

type Hub struct {
	Broadcast        chan []byte
	Register         chan *Client
	Unregister       chan *Client
	Clients          map[string]*Client // Map by client ID (UUID) for all clients
	UserClients      map[int]*Client    // Map by UserID for authenticated users (for backward compatibility)
	incomingMessages chan clientMessage
	editingSessions  map[string]*Client
	sessionCommands  chan sessionCommand
	authRepo         *repository.AuthRepository
}

type EditingSession struct {
	UserID    int
	Entity    string
	ContextID int
}

type sessionCommand struct {
	action     string
	client     *Client
	sessionKey string
	payload    map[string]interface{}
}

type clientMessage struct {
	client  *Client
	message []byte
}

func NewHub(authRepo *repository.AuthRepository) *Hub {
	return &Hub{
		Broadcast:        make(chan []byte),
		Register:         make(chan *Client),
		Unregister:       make(chan *Client),
		Clients:          make(map[string]*Client),
		UserClients:      make(map[int]*Client),
		incomingMessages: make(chan clientMessage),
		editingSessions:  make(map[string]*Client),
		sessionCommands:  make(chan sessionCommand),
		authRepo:         authRepo,
	}
}

// RUN HUB AS GOROUTINE
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// Register client by unique ID (supports multiple public users)
			h.Clients[client.ID] = client
			
			// For authenticated users (UserID > 0), also store in UserClients map
			// If an old connection exists for this user, close it
			if client.UserID > 0 {
				if oldClient, ok := h.UserClients[client.UserID]; ok {
					close(oldClient.Send)
					delete(h.Clients, oldClient.ID)
				}
				h.UserClients[client.UserID] = client
				log.Printf("WebSocket client registered: ClientID=%s, UserID=%d (authenticated)", client.ID, client.UserID)
			} else {
				log.Printf("WebSocket client registered: ClientID=%s, UserID=%d (public/anonymous)", client.ID, client.UserID)
			}
			
		case client := <-h.Unregister:
			// Remove from Clients map
			if _, ok := h.Clients[client.ID]; ok {
				h.cleanupClientSessions(client)
				delete(h.Clients, client.ID)
				close(client.Send)
				
				// For authenticated users, also remove from UserClients map
				if client.UserID > 0 {
					if registeredClient, ok := h.UserClients[client.UserID]; ok && registeredClient.ID == client.ID {
						delete(h.UserClients, client.UserID)
					}
					log.Printf("WebSocket client unregistered: ClientID=%s, UserID=%d (authenticated)", client.ID, client.UserID)
				} else {
					log.Printf("WebSocket client unregistered: ClientID=%s, UserID=%d (public/anonymous)", client.ID, client.UserID)
				}
			}
			
		case message := <-h.Broadcast:
			// Broadcast to all clients (both authenticated and public)
			for clientID, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, clientID)
					if client.UserID > 0 {
						delete(h.UserClients, client.UserID)
					}
				}
			}
			
		case clientMsg := <-h.incomingMessages:
			h.handleIncomingMessage(clientMsg.client, clientMsg.message)
			
		case cmd := <-h.sessionCommands:
			switch cmd.action {
			case "start":
				if _, exists := h.editingSessions[cmd.sessionKey]; !exists {
					h.editingSessions[cmd.sessionKey] = cmd.client
					log.Printf("User %d started editing %s", cmd.client.UserID, cmd.sessionKey)
					broadcastMsg, _ := NewMessage("EDITING_STARTED", cmd.payload)
					h.broadcastToOthers(broadcastMsg, cmd.client)
				}
			case "finish":
				if holder, exists := h.editingSessions[cmd.sessionKey]; exists && holder == cmd.client {
					delete(h.editingSessions, cmd.sessionKey)
					log.Printf("User %d finished editing %s", cmd.client.UserID, cmd.sessionKey)
					broadcastMsg, _ := NewMessage("EDITING_FINISHED", cmd.payload)
					h.broadcastToOthers(broadcastMsg, cmd.client)
				}
			}
		}
	}
}

func (h *Hub) BroadcastMessage(message []byte) {
	h.Broadcast <- message
}

func (h *Hub) handleIncomingMessage(client *Client, rawMessage []byte) {
	var msg Message
	if err := json.Unmarshal(rawMessage, &msg); err != nil {
		log.Printf("Error unmarshalling incoming message: %v", err)
		return
	}

	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		log.Printf("Invalid payload format for event: %s", msg.Event)
		return
	}

	entity, ok1 := payload["entity"].(string)
	contextIDFloat, ok2 := payload["context_id"].(float64)
	if !ok1 || !ok2 {
		log.Printf("Payload missing 'entity' or 'context_id' for event: %s", msg.Event)
		return
	}
	contextID := int(contextIDFloat)
	sessionKey := fmt.Sprintf("%s:%d", entity, contextID)

	switch msg.Event {
	case "START_EDITING":
		// Only authenticated users can start editing sessions
		if client.UserID == 0 {
			log.Printf("Public user attempted to start editing session - rejected")
			return
		}

		if _, exists := h.editingSessions[sessionKey]; !exists {
			h.editingSessions[sessionKey] = client
			log.Printf("User %d started editing %s", client.UserID, sessionKey)

			broadcastMsg, err := NewMessage("EDITING_STARTED", payload)

			if err != nil {
				log.Printf("CRITICAL: Failed to create broadcast message for start reorder: %v", err)
			} else {
				h.broadcastToOthers(broadcastMsg, client)
			}
		}

	case "FINISH_EDITING":
		// Only authenticated users can finish editing sessions
		if client.UserID == 0 {
			log.Printf("Public user attempted to finish editing session - rejected")
			return
		}

		if holder, exists := h.editingSessions[sessionKey]; exists && holder == client {
			delete(h.editingSessions, sessionKey)
			log.Printf("User %d finished editing %s", client.UserID, sessionKey)

			broadcastMsg, err := NewMessage("EDITING_FINISHED", payload)

			if err != nil {
				log.Printf("CRITICAL: Failed to create broadcast message for finish reorder: %v", err)
			} else {
				h.broadcastToOthers(broadcastMsg, client)
			}

		}
	}
}

func (h *Hub) cleanupClientSessions(client *Client) {
	for key, holder := range h.editingSessions {
		if holder == client {
			delete(h.editingSessions, key)
			log.Printf("Cleaned up editing session %s for disconnected user %d", key, client.UserID)
		}
	}
}

func (h *Hub) broadcastToOthers(message []byte, exclude *Client) {
	for _, client := range h.Clients {
		if client.ID != exclude.ID {
			client.Send <- message
		}
	}
}

func (h *Hub) ReleaseLock(client *Client, entity string, contextID int) {
	sessionKey := fmt.Sprintf("%s:%d", entity, contextID)

	if holder, exists := h.editingSessions[sessionKey]; exists && holder == client {
		delete(h.editingSessions, sessionKey)
		log.Printf("User %d released lock for %s", client.UserID, sessionKey)

		payload := map[string]interface{}{
			"entity":     entity,
			"context_id": contextID,
		}
		broadcastMsg, err := NewMessage("EDITING_FINISHED", payload)

		if err != nil {
			log.Printf("CRITICAL: Failed to create websocket message for finish reorder: %v", err)
		} else {
			h.broadcastToOthers(broadcastMsg, client)
		}

	}
}

func (h *Hub) GetClientByUserID(userID int) *Client {
	if client, ok := h.UserClients[userID]; ok {
		return client
	}
	return nil
}

func (h *Hub) SendConnectionEstablished(client *Client) {
	isEditing, err := h.authRepo.GetEditMode(context.Background())
	if err != nil {
		log.Printf("Error getting edit mode for new client: %v", err)
	}

	payload := gin.H{
		"system_status": gin.H{
			"is_editing": isEditing,
		},
	}
	message, err := NewMessage("CONNECTION_ESTABLISHED", payload)
	if err != nil {
		log.Printf("CRITICAL: Failed to create connection established message: %v", err)
		return
	}

	client.Send <- message
}

package handler

import (
	"log"
	"net/http"
	"os"
	"strings"

	"e-memo-job-reservation-api/internal/repository"
	ws "e-memo-job-reservation-api/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
		origin := r.Header.Get("Origin")
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
		// TODO: ADJUST ALLOWED ORIGIN WITH PRODUCTION IP ADDRESS
		return true
	},
}

type WebSocketHandler struct {
	hub      *ws.Hub
	authRepo *repository.AuthRepository
}

func NewWebSocketHandler(hub *ws.Hub, authRepo *repository.AuthRepository) *WebSocketHandler {
	return &WebSocketHandler{hub: hub, authRepo: authRepo}
}

// HANDLE WEBSOCKET REQUEST FROM CLIENT
func (h *WebSocketHandler) ServeWs(c *gin.Context) {
	ticket := c.Query("ticket")
	if ticket == "" {
		return
	}

	userID, err := h.authRepo.ValidateAndDelWebSocketTicket(c.Request.Context(), ticket)
	if err != nil {
		log.Printf("WebSocket connection rejected: %v", err)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v", err)
		return
	}

	// Generate unique client ID
	clientID := uuid.New().String()

	client := &ws.Client{
		ID:     clientID,
		Hub:    h.hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: userID,
	}
	client.Hub.Register <- client

	go h.hub.SendConnectionEstablished(client)

	go client.WritePump()
	go client.ReadPump()
}

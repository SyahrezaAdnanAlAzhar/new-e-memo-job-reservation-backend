package handler

import (
	"log"
	"net/http"
	"strings"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/service"
	"e-memo-job-reservation-api/internal/util"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{Service: service}
}

type LoginRequest struct {
	NPK string `json:"npk" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Username and password are required", nil)
		return
	}

	loginResponse, err := h.Service.Login(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "invalid credentials" {
			util.ErrorResponse(c, http.StatusUnauthorized, "Invalid username or password", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to login", err.Error())
		return
	}

	util.SuccessResponse(c, http.StatusOK, loginResponse)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Refresh token is required", nil)
		return
	}

	loginResponse, err := h.Service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "not found") {
			util.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired refresh token", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to process refresh token", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, loginResponse)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		util.ErrorResponse(c, http.StatusBadRequest, "Authorization header is required", nil)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid token format", nil)
		return
	}
	tokenString := parts[1]

	err := h.Service.Logout(c.Request.Context(), tokenString)
	if err != nil {
		log.Printf("Error during logout process: %v", err)
		// logout failure usually not blocking response, so proceed
	}

	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// POST /auth/ws-ticket
func (h *AuthHandler) GenerateWebSocketTicket(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		util.ErrorResponse(c, http.StatusUnauthorized, "User ID not found in token", nil)
		return
	}

	ticket, err := h.Service.GenerateWebSocketTicket(c.Request.Context(), userID.(int))
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate WebSocket ticket", err.Error())
		return
	}

	util.SuccessResponse(c, http.StatusOK, gin.H{"ticket": ticket})
}

// POST /auth/ws-public-ticket
func (h *AuthHandler) GeneratePublicWebSocketTicket(c *gin.Context) {
	log.Println("[DEBUG] Generating public WebSocket ticket...")
	ticket, err := h.Service.GeneratePublicWebSocketTicket(c.Request.Context())
	if err != nil {
		log.Printf("[ERROR] Failed to generate public WebSocket ticket: %v", err)
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate public WebSocket ticket", err.Error())
		return
	}

	log.Printf("[DEBUG] Successfully generated public WebSocket ticket: %s", ticket)
	util.SuccessResponse(c, http.StatusOK, gin.H{"ticket": ticket})
}

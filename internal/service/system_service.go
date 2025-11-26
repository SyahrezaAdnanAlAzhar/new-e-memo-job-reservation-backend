package service

import (
	"context"
	"log"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/repository"
	"e-memo-job-reservation-api/internal/websocket"

	"github.com/gin-gonic/gin"
)

type SystemService struct {
	authRepo *repository.AuthRepository
	hub      *websocket.Hub
}

func NewSystemService(authRepo *repository.AuthRepository, hub *websocket.Hub) *SystemService {
	return &SystemService{
		authRepo: authRepo,
		hub:      hub,
	}
}

func (s *SystemService) UpdateEditMode(ctx context.Context, req dto.UpdateEditModeRequest) error {
	err := s.authRepo.SetEditMode(ctx, req.IsEditing)
	if err != nil {
		return err
	}

	payload := gin.H{
		"is_editing": req.IsEditing,
		"message":    "System edit mode has been updated.",
	}
	message, err := websocket.NewMessage("SYSTEM_EDIT_MODE_CHANGED", payload)
	if err != nil {
		log.Printf("CRITICAL: Failed to create websocket message for edit mode change: %v", err)
	} else {
		s.hub.BroadcastMessage(message)
	}

	return nil
}

package service

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/repository"
	"e-memo-job-reservation-api/internal/websocket"

	"github.com/gin-gonic/gin"
)

type TicketPriorityService struct {
	db           *sql.DB
	hub          *websocket.Hub
	ticketRepo   *repository.TicketRepository
	employeeRepo *repository.EmployeeRepository
}

func NewTicketPriorityService(db *sql.DB, hub *websocket.Hub, ticketRepo *repository.TicketRepository, employeeRepo *repository.EmployeeRepository) *TicketPriorityService {
	return &TicketPriorityService{
		db:           db,
		hub:          hub,
		ticketRepo:   ticketRepo,
		employeeRepo: employeeRepo,
	}
}

// RE ORDER
func (s *TicketPriorityService) ReorderTickets(ctx context.Context, req dto.ReorderTicketsRequest, userNPK string) error {
	_, err := s.employeeRepo.FindByNPK(userNPK)
	if err != nil {
		return errors.New("action performer not found")
	}

	ticketIDs := make([]int, len(req.Items))
	for i, item := range req.Items {
		ticketIDs[i] = item.TicketID
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, item := range req.Items {
		newPriority := i + 1
		rowsAffected, err := s.ticketRepo.UpdatePriority(ctx, tx, item.TicketID, item.Version, newPriority)
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return errors.New("data conflict: one or more tickets have been modified by another user, please refresh")
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	payload := gin.H{
		"department_target_id": req.DepartmentTargetID,
		"message":              "Ticket priorities have been updated.",
	}
	message, err := websocket.NewMessage("TICKET_PRIORITY_UPDATED", payload)
	if err != nil {
		log.Printf("CRITICAL: Failed to create websocket message for ticket reorder: %v", err)
	} else {
		s.hub.BroadcastMessage(message)
	}

	return nil
}

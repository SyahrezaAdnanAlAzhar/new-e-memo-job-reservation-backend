package service

import (
	"context"
	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type StatusTicketService struct {
	repo *repository.StatusTicketRepository
}

func NewStatusTicketService(repo *repository.StatusTicketRepository) *StatusTicketService {
	return &StatusTicketService{repo: repo}
}

// CREATE
func (s *StatusTicketService) CreateStatusTicket(req dto.CreateStatusTicketRequest) (*model.StatusTicket, error) {
	newStatus, err := s.repo.Create(req)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New("status ticket name or sequence already exists")
		}
		return nil, err
	}
	return newStatus, nil
}

// GET ALL
func (s *StatusTicketService) GetAllStatusTickets(filters dto.StatusTicketFilter) ([]model.StatusTicket, error) {
	return s.repo.FindAll(filters)
}

// GET BY ID
func (s *StatusTicketService) GetStatusTicketByID(id int) (*model.StatusTicket, error) {
	return s.repo.FindByID(id)
}

// DELETE
func (s *StatusTicketService) DeleteStatusTicket(id int) error {
	return s.repo.Delete(id)
}

// CHANGE STATUS
func (s *StatusTicketService) UpdateStatusTicketActiveStatus(id int, req dto.UpdateStatusTicketStatusRequest) error {
	return s.repo.UpdateActiveStatus(id, req.IsActive)
}

// REORDER
func (s *StatusTicketService) ReorderStatusTickets(req dto.ReorderStatusTicketsRequest) error {
	ctx := context.Background()
	tx, err := s.repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// Delete Section (Sequence <= -100)
	// -100, -101, -102...
	for i, id := range req.DeleteSectionOrder {
		newSequence := -100 - i
		if err := s.repo.Reorder(ctx, tx, id, newSequence); err != nil {
			return err
		}
	}

	// Approval Section (-99 < Sequence < 0)
	// -1, -2, -3...
	for i, id := range req.ApprovalSectionOrder {
		newSequence := -1 - i
		if err := s.repo.Reorder(ctx, tx, id, newSequence); err != nil {
			return err
		}
	}

	// Actual Section (Sequence >= 0)
	// 0, 1, 2...
	for i, id := range req.ActualSectionOrder {
		newSequence := i
		if err := s.repo.Reorder(ctx, tx, id, newSequence); err != nil {
			return err
		}
	}

	return tx.Commit()
}

package service

import (
	"context"
	"database/sql"
	"errors"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
)

type EmployeePositionService struct {
	positionRepo     *repository.EmployeePositionRepository
	mappingRepo      *repository.PositionToWorkflowMappingRepository
	ticketRepo       *repository.TicketRepository
	statusTicketRepo *repository.StatusTicketRepository
	db               *sql.DB
}

func NewEmployeePositionService(
	positionRepo *repository.EmployeePositionRepository,
	mappingRepo *repository.PositionToWorkflowMappingRepository,
	ticketRepo *repository.TicketRepository,
	statusTicketRepo *repository.StatusTicketRepository,
	db *sql.DB) *EmployeePositionService {
	return &EmployeePositionService{positionRepo: positionRepo, mappingRepo: mappingRepo, ticketRepo: ticketRepo, statusTicketRepo: statusTicketRepo, db: db}
}

// CREATE
func (s *EmployeePositionService) CreateEmployeePosition(ctx context.Context, req dto.CreateEmployeePositionRequest) (*model.EmployeePosition, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	newPos, err := s.positionRepo.Create(ctx, tx, req)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New("position name already exists")
		}
		return nil, err
	}

	err = s.mappingRepo.CreateWorkflowMapping(ctx, tx, newPos.ID, req.WorkflowID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, errors.New("invalid workflow_id")
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return newPos, nil
}

// GET ALL
func (s *EmployeePositionService) GetAllEmployeePositions() ([]model.EmployeePosition, error) {
	return s.positionRepo.FindAll()
}

// GET BY ID
func (s *EmployeePositionService) GetEmployeePositionByID(id int) (*model.EmployeePosition, error) {
	return s.positionRepo.FindByID(id)
}

// UPDATE
func (s *EmployeePositionService) UpdateEmployeePosition(id int, req dto.UpdateEmployeePositionRequest) (*model.EmployeePosition, error) {
	isTaken, err := s.positionRepo.IsNameTaken(req.Name, id)
	if err != nil {
		return nil, err
	}
	if isTaken {
		return nil, errors.New("position name already exists")
	}
	return s.positionRepo.Update(id, req)
}

// DELETE
func (s *EmployeePositionService) DeleteEmployeePosition(ctx context.Context, id int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	cancelledStatus, err := s.statusTicketRepo.FindBySequence(0)
	if err != nil {
		return errors.New("critical configuration error: 'Dibatalkan' status not found")
	}

	if err := s.ticketRepo.CancelTicketsByPosition(ctx, tx, id, cancelledStatus.ID); err != nil {
		return err
	}

	if err := s.positionRepo.Delete(ctx, tx, id); err != nil {
		return err
	}

	return tx.Commit()
}

// CHANGE STATUS
func (s *EmployeePositionService) UpdateEmployeePositionActiveStatus(id int, req dto.UpdateEmployeePositionStatusRequest) error {
	return s.positionRepo.UpdateActiveStatus(id, req.IsActive)
}

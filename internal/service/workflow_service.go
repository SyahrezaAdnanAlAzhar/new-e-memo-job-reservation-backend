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

type WorkflowService struct {
	workflowRepo *repository.WorkflowRepository
	stepRepo     *repository.WorkflowStepRepository
	db           *sql.DB
}

func NewWorkflowService(workflowRepo *repository.WorkflowRepository, stepRepo *repository.WorkflowStepRepository, db *sql.DB) *WorkflowService {
	return &WorkflowService{workflowRepo: workflowRepo, stepRepo: stepRepo, db: db}
}

// CREATE
func (s *WorkflowService) CreateWorkflowWithSteps(ctx context.Context, req dto.CreateWorkflowRequest) (*model.Workflow, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// CREATE WORKFLOW
	newWorkflow, err := s.workflowRepo.Create(ctx, tx, req.Name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New("workflow name already exists")
		}
		return nil, err
	}

	// LOOP FOR EACH WORKFLOW STEP
	for i, statusID := range req.StatusTicketIDs {
		stepSequence := i
		err := s.stepRepo.Create(ctx, tx, newWorkflow.ID, statusID, stepSequence)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23503" { // foreign_key_violation
					return nil, errors.New("one or more status_ticket_ids are invalid")
				}
				if pgErr.Code == "23505" { // unique_violation
					return nil, errors.New("cannot add the same status twice to a workflow")
				}
			}
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return newWorkflow, nil
}

// CREATE WORKFLOW STEP
func (s *WorkflowService) AddWorkflowStep(ctx context.Context, req dto.AddWorkflowStepRequest) (*model.WorkflowStep, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var newSequence int
	if req.Position == "start" {
		if err := s.stepRepo.IncrementAllSequences(ctx, tx, req.WorkflowID); err != nil {
			return nil, err
		}
		newSequence = 0
	} else {
		lastSequence, err := s.stepRepo.GetLastSequence(ctx, tx, req.WorkflowID)
		if err != nil {
			return nil, err
		}
		newSequence = lastSequence + 1
	}

	if err := s.stepRepo.Create(ctx, tx, req.WorkflowID, req.StatusTicketID, newSequence); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New("status ticket is already in this workflow")
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return nil, nil
}

// GET ALL
func (s *WorkflowService) GetAllWorkflows() ([]model.Workflow, error) {
	return s.workflowRepo.FindAll()
}

func (s *WorkflowService) GetAllWorkflowSteps() ([]model.WorkflowStep, error) {
	return s.stepRepo.FindAll()
}

// GET BY ID
func (s *WorkflowService) GetWorkflowByID(id int) (*model.Workflow, error) {
	return s.workflowRepo.FindByID(id)
}

func (s *WorkflowService) GetWorkflowStepsByWorkflowID(workflowID int) ([]model.WorkflowStep, error) {
	return s.stepRepo.FindByWorkflowID(workflowID)
}

func (s *WorkflowService) GetWorkflowStepByID(id int) (*model.WorkflowStep, error) {
	return s.stepRepo.FindByID(id)
}

// UPDATE WORKFLOW NAME
func (s *WorkflowService) UpdateWorkflowName(id int, req dto.UpdateWorkflowRequest) (*model.Workflow, error) {
	isTaken, err := s.workflowRepo.IsNameTaken(req.Name, id)
	if err != nil {
		return nil, err
	}
	if isTaken {
		return nil, errors.New("workflow name already exists")
	}
	return s.workflowRepo.Update(id, req)
}

// DELETE WORKFLOW
func (s *WorkflowService) DeleteWorkflow(id int) error {
	return s.workflowRepo.Delete(id)
}

// DELETE WORKFLOW STEP
func (s *WorkflowService) DeleteWorkflowStep(id int) error {
	return s.stepRepo.Delete(id)
}

// CHANGE WORKFLOW STATUS
func (s *WorkflowService) UpdateWorkflowActiveStatus(ctx context.Context, id int, req dto.UpdateWorkflowStatusRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.workflowRepo.UpdateActiveStatus(ctx, tx, id, req.IsActive); err != nil {
		return err
	}

	if err := s.stepRepo.UpdateActiveStatusByWorkflowID(ctx, tx, id, req.IsActive); err != nil {
		return err
	}

	return tx.Commit()
}

// CHANGE WORKFLOW STEP STATUS
func (s *WorkflowService) UpdateWorkflowStepActiveStatus(id int, req dto.UpdateWorkflowStatusRequest) error {
	return s.stepRepo.UpdateActiveStatus(id, req.IsActive)
}

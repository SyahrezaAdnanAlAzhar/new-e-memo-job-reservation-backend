package repository

import (
	"context"
	"database/sql"
	"e-memo-job-reservation-api/internal/model"
)

type WorkflowStepRepository struct {
	DB *sql.DB
}

func NewWorkflowStepRepository(db *sql.DB) *WorkflowStepRepository {
	return &WorkflowStepRepository{DB: db}
}

// CREATE
func (r *WorkflowStepRepository) Create(ctx context.Context, tx *sql.Tx, workflowID, statusTicketID, stepSequence int) error {
	query := "INSERT INTO workflow_step (workflow_id, status_ticket_id, step_sequence, is_active) VALUES ($1, $2, $3, false)"
	_, err := tx.ExecContext(ctx, query, workflowID, statusTicketID, stepSequence)
	return err
}

// GET ALL
func (r *WorkflowStepRepository) FindAll() ([]model.WorkflowStep, error) {
	query := "SELECT id, workflow_id, status_ticket_id, step_sequence, is_active, created_at, updated_at FROM workflow_step ORDER BY workflow_id, step_sequence ASC"
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []model.WorkflowStep
	for rows.Next() {
		var s model.WorkflowStep
		err := rows.Scan(&s.ID, &s.WorkflowID, &s.StatusTicketID, &s.StepSequence, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}
	return steps, nil
}

// GET ALL BY WORKFLOW ID
func (r *WorkflowStepRepository) FindByWorkflowID(workflowID int) ([]model.WorkflowStep, error) {
	query := "SELECT id, workflow_id, status_ticket_id, step_sequence, is_active, created_at, updated_at FROM workflow_step WHERE workflow_id = $1 ORDER BY step_sequence ASC"
	rows, err := r.DB.Query(query, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []model.WorkflowStep
	for rows.Next() {
		var s model.WorkflowStep
		err := rows.Scan(&s.ID, &s.WorkflowID, &s.StatusTicketID, &s.StepSequence, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}
	return steps, nil
}

// GET BY ID
func (r *WorkflowStepRepository) FindByID(id int) (*model.WorkflowStep, error) {
	query := "SELECT id, workflow_id, status_ticket_id, step_sequence, is_active, created_at, updated_at FROM workflow_step WHERE id = $1"
	row := r.DB.QueryRow(query, id)
	var s model.WorkflowStep
	err := row.Scan(&s.ID, &s.WorkflowID, &s.StatusTicketID, &s.StepSequence, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

// HELPER FOR INSERT STEP
func (r *WorkflowStepRepository) GetLastSequence(ctx context.Context, tx *sql.Tx, workflowID int) (int, error) {
	var lastSequence sql.NullInt64
	query := "SELECT MAX(step_sequence) FROM workflow_step WHERE workflow_id = $1"
	err := tx.QueryRowContext(ctx, query, workflowID).Scan(&lastSequence)
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}
	if !lastSequence.Valid {
		return -1, nil
	}
	return int(lastSequence.Int64), nil
}

// Shift all sequences up to make room early
func (r *WorkflowStepRepository) IncrementAllSequences(ctx context.Context, tx *sql.Tx, workflowID int) error {
	query := "UPDATE workflow_step SET step_sequence = step_sequence + 1 WHERE workflow_id = $1"
	_, err := tx.ExecContext(ctx, query, workflowID)
	return err
}

// DELETE
func (r *WorkflowStepRepository) Delete(id int) error {
	query := "DELETE FROM workflow_step WHERE id = $1 AND step_sequence = 0"
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CHANGE STATUS
func (r *WorkflowStepRepository) UpdateActiveStatus(id int, isActive bool) error {
	query := "UPDATE workflow_step SET is_active = $1, updated_at = NOW() WHERE id = $2"
	result, err := r.DB.Exec(query, isActive, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CHANGE STATUS (CASCADE DARI WORKFLOW)
func (r *WorkflowStepRepository) UpdateActiveStatusByWorkflowID(ctx context.Context, tx *sql.Tx, workflowID int, isActive bool) error {
	query := "UPDATE workflow_step SET is_active = $1, updated_at = NOW() WHERE workflow_id = $2"
	_, err := tx.ExecContext(ctx, query, isActive, workflowID)
	return err
}

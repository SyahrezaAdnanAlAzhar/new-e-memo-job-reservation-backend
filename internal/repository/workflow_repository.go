package repository

import (
	"context"
	"database/sql"
	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
)

type WorkflowRepository struct {
	DB *sql.DB
}

func NewWorkflowRepository(db *sql.DB) *WorkflowRepository {
	return &WorkflowRepository{DB: db}
}

// CREATE
func (r *WorkflowRepository) Create(ctx context.Context, tx *sql.Tx, name string) (*model.Workflow, error) {
	query := `
        INSERT INTO workflow (name, is_active) VALUES ($1, false)
        RETURNING id, name, is_active, created_at, updated_at`

	row := tx.QueryRowContext(ctx, query, name)

	var newWorkflow model.Workflow
	err := row.Scan(
		&newWorkflow.ID, &newWorkflow.Name, &newWorkflow.IsActive,
		&newWorkflow.CreatedAt, &newWorkflow.UpdatedAt,
	)
	return &newWorkflow, err
}

// GET ALL
func (r *WorkflowRepository) FindAll() ([]model.Workflow, error) {
	query := "SELECT id, name, is_active, created_at, updated_at FROM workflow ORDER BY id ASC"
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []model.Workflow
	for rows.Next() {
		var w model.Workflow
		err := rows.Scan(&w.ID, &w.Name, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
		if err != nil {
			return nil, err
		}
		workflows = append(workflows, w)
	}
	return workflows, nil
}

// GET BY ID
func (r *WorkflowRepository) FindByID(id int) (*model.Workflow, error) {
	query := "SELECT id, name, is_active, created_at, updated_at FROM workflow WHERE id = $1"
	row := r.DB.QueryRow(query, id)
	var w model.Workflow
	err := row.Scan(&w.ID, &w.Name, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	return &w, err
}

// UPDATE
func (r *WorkflowRepository) Update(id int, req dto.UpdateWorkflowRequest) (*model.Workflow, error) {
	query := `UPDATE workflow SET name = $1, updated_at = NOW() WHERE id = $2
              RETURNING id, name, is_active, created_at, updated_at`
	row := r.DB.QueryRow(query, req.Name, id)
	var updatedWorkflow model.Workflow
	err := row.Scan(&updatedWorkflow.ID, &updatedWorkflow.Name, &updatedWorkflow.IsActive, &updatedWorkflow.CreatedAt, &updatedWorkflow.UpdatedAt)
	return &updatedWorkflow, err
}

// DELETE
func (r *WorkflowRepository) Delete(id int) error {
	query := "DELETE FROM workflow WHERE id = $1"
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
func (r *WorkflowRepository) UpdateActiveStatus(ctx context.Context, tx *sql.Tx, id int, isActive bool) error {
	query := "UPDATE workflow SET is_active = $1, updated_at = NOW() WHERE id = $2"
	result, err := tx.ExecContext(ctx, query, isActive, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GET NEXT WORKFLOW STEP
func (r *WorkflowRepository) GetNextWorkflowStep(ctx context.Context, currentStatusID int) (nextStatusID int, isFinalStep bool, err error) {
	query := `
        WITH current_step AS (
            SELECT workflow_id, step_sequence
            FROM workflow_step
            WHERE status_ticket_id = $1
        )
        SELECT ws.status_ticket_id
        FROM workflow_step ws
        WHERE ws.workflow_id = (SELECT workflow_id FROM current_step)
          AND ws.step_sequence = (SELECT step_sequence FROM current_step) + 1
        LIMIT 1`

	err = r.DB.QueryRowContext(ctx, query, currentStatusID).Scan(&nextStatusID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, true, nil
		}
		return 0, false, err
	}

	return nextStatusID, false, nil
}

// HELPER
// GET INITIAL STATUS BY RULE
func (r *WorkflowRepository) GetInitialStatusByPosition(ctx context.Context, positionID int) (int, error) {
	var statusID int
	query := `
        SELECT ws.status_ticket_id
        FROM workflow_step ws
        JOIN position_to_workflow_mapping ptwm ON ws.workflow_id = ptwm.workflow_id
        WHERE ptwm.employee_position_id = $1
		ORDER BY ws.step_sequence ASC
        LIMIT 1`

	err := r.DB.QueryRowContext(ctx, query, positionID).Scan(&statusID)
	if err != nil {
		return 0, err
	}
	return statusID, nil
}

// UPDATE VALIDATION
func (r *WorkflowRepository) IsNameTaken(name string, currentID int) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM workflow WHERE name = $1 AND id != $2)"
	err := r.DB.QueryRow(query, name, currentID).Scan(&exists)
	return exists, err
}

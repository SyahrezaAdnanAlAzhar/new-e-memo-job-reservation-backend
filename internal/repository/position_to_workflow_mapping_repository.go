package repository

import (
	"context"
	"database/sql"
	// "e-memo-job-reservation-api/internal/dto"
	// "e-memo-job-reservation-api/internal/model"
)

type PositionToWorkflowMappingRepository struct {
	DB *sql.DB
}

func NewPositionToWorkflowMappingRepository(db *sql.DB) *PositionToWorkflowMappingRepository {
	return &PositionToWorkflowMappingRepository{DB: db}
}

// CREATE WORKFLOW MAPPING
func (r *PositionToWorkflowMappingRepository) CreateWorkflowMapping(ctx context.Context, tx *sql.Tx, positionID int, workflowID int) error {
	query := "INSERT INTO position_to_workflow_mapping (employee_position_id, workflow_id) VALUES ($1, $2)"
	_, err := tx.ExecContext(ctx, query, positionID, workflowID)
	return err
}

package repository

import (
	"context"
	"database/sql"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
)

type TicketActionLogRepository struct {
	DB *sql.DB
}

func NewTicketActionLogRepository(db *sql.DB) *TicketActionLogRepository {
	return &TicketActionLogRepository{DB: db}
}

func (r *TicketActionLogRepository) Create(ctx context.Context, tx *sql.Tx, logEntry model.TicketActionLog) error {
	query := `
        INSERT INTO ticket_action_log (
            ticket_id, action_id, performed_by_npk, details_text, 
            file_path, from_status_id, to_status_id
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := tx.ExecContext(ctx, query,
		logEntry.TicketID,
		logEntry.ActionID,
		logEntry.PerformedByNpk,
		logEntry.DetailsText,
		logEntry.FilePath,
		logEntry.FromStatusID,
		logEntry.ToStatusID,
	)
	return err
}

func (r *TicketActionLogRepository) FindLastRejectionByTicketID(ctx context.Context, ticketID int) (*dto.RejectionDetailResponse, error) {
	query := `
        SELECT 
            tal.details_text,
            tal.performed_by_npk,
            e.name as rejector_name,
            ep.name as rejector_position,
            d.name as rejector_department,
            tal.performed_at
        FROM ticket_action_log tal
        JOIN action a ON tal.action_id = a.id
        JOIN employee e ON tal.performed_by_npk = e.npk
        JOIN employee_position ep ON e.employee_position_id = ep.id
        JOIN department d ON e.department_id = d.id
        WHERE tal.ticket_id = $1
          AND a.name IN ('Tolak', 'Tolak Hasil Job')
        ORDER BY tal.performed_at DESC
        LIMIT 1`

	row := r.DB.QueryRowContext(ctx, query, ticketID)

	var rejectionDetail dto.RejectionDetailResponse
	err := row.Scan(
		&rejectionDetail.Reason,
		&rejectionDetail.RejectorNPK,
		&rejectionDetail.RejectorName,
		&rejectionDetail.RejectorPosition,
		&rejectionDetail.RejectorDepartment,
		&rejectionDetail.RejectedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rejectionDetail, nil
}

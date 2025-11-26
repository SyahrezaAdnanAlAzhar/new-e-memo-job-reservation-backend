package repository

import (
	"context"
	"database/sql"
	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
)

type RejectedTicketRepository struct {
	DB *sql.DB
}

func NewRejectedTicketRepository(db *sql.DB) *RejectedTicketRepository {
	return &RejectedTicketRepository{DB: db}
}

// CREATE
func (r *RejectedTicketRepository) Create(ctx context.Context, tx *sql.Tx, req dto.CreateRejectedTicketRequest, rejectorNPK string) (*model.RejectedTicket, error) {
	query := `
        INSERT INTO rejected_ticket (ticket_id, rejector, feedback, already_seen) 
        VALUES ($1, $2, $3, false)
        RETURNING id, ticket_id, rejector, feedback, already_seen, created_at, updated_at`

	row := tx.QueryRowContext(ctx, query, req.TicketID, rejectorNPK, req.Feedback)

	var newRejection model.RejectedTicket
	err := row.Scan(
		&newRejection.ID, &newRejection.TicketID, &newRejection.Rejector, &newRejection.Feedback,
		&newRejection.AlreadySeen, &newRejection.CreatedAt, &newRejection.UpdatedAt,
	)
	return &newRejection, err
}

// UPDATE FEEDBACK
func (r *RejectedTicketRepository) UpdateFeedback(id int64, feedback string) (*model.RejectedTicket, error) {
	query := `
        UPDATE rejected_ticket SET feedback = $1, updated_at = NOW() WHERE id = $2
        RETURNING id, ticket_id, rejector, feedback, already_seen, created_at, updated_at`

	row := r.DB.QueryRow(query, feedback, id)

	var updatedRejection model.RejectedTicket
	err := row.Scan(
		&updatedRejection.ID, &updatedRejection.TicketID, &updatedRejection.Rejector, &updatedRejection.Feedback,
		&updatedRejection.AlreadySeen, &updatedRejection.CreatedAt, &updatedRejection.UpdatedAt,
	)
	return &updatedRejection, err
}

// UPDATE ALREADY_SEEN
func (r *RejectedTicketRepository) UpdateAlreadySeen(id int64, seen bool) error {
	query := "UPDATE rejected_ticket SET already_seen = $1, updated_at = NOW() WHERE id = $2"
	result, err := r.DB.Exec(query, seen, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DELETE
func (r *RejectedTicketRepository) Delete(id int64) error {
	query := "DELETE FROM rejected_ticket WHERE id = $1"
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

// HELPER: Find Latest By Ticket ID
func (r *RejectedTicketRepository) FindLatestByTicketID(ctx context.Context, ticketID int64) (*model.RejectedTicket, error) {
	query := "SELECT id, ticket_id, rejector, feedback, already_seen FROM rejected_ticket WHERE ticket_id = $1 ORDER BY created_at DESC LIMIT 1"
	row := r.DB.QueryRowContext(ctx, query, ticketID)
	var rejection model.RejectedTicket
	err := row.Scan(&rejection.ID, &rejection.TicketID, &rejection.Rejector, &rejection.Feedback, &rejection.AlreadySeen)
	return &rejection, err
}

// HELPER: Find By ID
func (r *RejectedTicketRepository) FindByID(id int64) (*model.RejectedTicket, error) {
	query := "SELECT id, ticket_id, rejector, feedback, already_seen FROM rejected_ticket WHERE id = $1"
	row := r.DB.QueryRow(query, id)
	var rejection model.RejectedTicket
	err := row.Scan(&rejection.ID, &rejection.TicketID, &rejection.Rejector, &rejection.Feedback, &rejection.AlreadySeen)
	return &rejection, err
}

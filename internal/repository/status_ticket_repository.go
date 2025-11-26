package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
)

type StatusTicketRepository struct {
	DB *sql.DB
}

func NewStatusTicketRepository(db *sql.DB) *StatusTicketRepository {
	return &StatusTicketRepository{DB: db}
}

// MAIN

// CREATE
func (r *StatusTicketRepository) Create(req dto.CreateStatusTicketRequest) (*model.StatusTicket, error) {
	query := `
        INSERT INTO status_ticket (name, sequence, is_active)
        VALUES ($1, $2, false)
        RETURNING id, name, sequence, is_active, created_at, updated_at`

	row := r.DB.QueryRow(query, req.Name, req.Sequence)

	var newStatus model.StatusTicket
	err := row.Scan(
		&newStatus.ID, &newStatus.Name, &newStatus.Sequence,
		&newStatus.IsActive, &newStatus.CreatedAt, &newStatus.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &newStatus, nil
}

// GET ALL
func (r *StatusTicketRepository) FindAll(filters dto.StatusTicketFilter) ([]model.StatusTicket, error) {
	baseQuery := "SELECT id, name, sequence, is_active, section_id, hex_color, created_at, updated_at FROM status_ticket"
	var conditions []string
	var args []interface{}
	argID := 1

	if filters.SectionID != 0 {
		conditions = append(conditions, fmt.Sprintf("section_id = $%d", argID))
		args = append(args, filters.SectionID)
		argID++
	}
	if filters.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argID))
		args = append(args, *filters.IsActive)
		argID++
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY sequence ASC"

	rows, err := r.DB.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []model.StatusTicket
	for rows.Next() {
		var s model.StatusTicket
		err := rows.Scan(&s.ID, &s.Name, &s.Sequence, &s.IsActive, &s.SectionID, &s.HexColor, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, s)
	}
	return statuses, nil
}

// GET BY ID
func (r *StatusTicketRepository) FindByID(id int) (*model.StatusTicket, error) {
	query := "SELECT id, name, sequence, is_active, created_at, updated_at FROM status_ticket WHERE id = $1"
	row := r.DB.QueryRow(query, id)

	var s model.StatusTicket
	err := row.Scan(&s.ID, &s.Name, &s.Sequence, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// DELETE
func (r *StatusTicketRepository) Delete(id int) error {
	query := "DELETE FROM status_ticket WHERE id = $1"
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

// CHANGE ACTIVE STATUS
func (r *StatusTicketRepository) UpdateActiveStatus(id int, isActive bool) error {
	query := "UPDATE status_ticket SET is_active = $1, updated_at = NOW() WHERE id = $2"
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

// REORDER
func (r *StatusTicketRepository) Reorder(ctx context.Context, tx *sql.Tx, id int, newSequence int) error {
	query := "UPDATE status_ticket SET sequence = $1, updated_at = NOW() WHERE id = $2"
	_, err := tx.ExecContext(ctx, query, newSequence, id)
	return err
}

// GET NEXT STATUS BASED ON SEQUENCE
func (r *StatusTicketRepository) GetNextStatusInSection(currentStatusID int) (*model.StatusTicket, error) {
	query := `
        WITH current_status AS (
            SELECT section_id, sequence
            FROM status_ticket
            WHERE id = $1
        )
        SELECT id, name, sequence, is_active
        FROM status_ticket
        WHERE section_id = (SELECT section_id FROM current_status)
          AND sequence > (SELECT sequence FROM current_status)
        ORDER BY sequence ASC
        LIMIT 1`

	var nextStatus model.StatusTicket
	err := r.DB.QueryRow(query, currentStatusID).Scan(&nextStatus.ID, &nextStatus.Name, &nextStatus.Sequence, &nextStatus.IsActive)
	if err != nil {
		return nil, err
	}
	return &nextStatus, nil
}

// GET SECTION FROM A STATUS
func (r *StatusTicketRepository) GetSectionID(statusID int) (int, error) {
	var sectionID int
	query := "SELECT section_id FROM status_ticket WHERE id = $1"
	err := r.DB.QueryRow(query, statusID).Scan(sectionID)
	return sectionID, err
}

// GET SECTION BY NAME
func (r *StatusTicketRepository) GetSectionIDByName(sectionName string) (int, error) {
	var sectionID int
	query := "SELECT id FROM section_status_ticket WHERE name = $1"
	err := r.DB.QueryRow(query, sectionName).Scan(sectionID)
	return sectionID, err
}

// CHANGE STATUS BASED ON SECTION
func (r *StatusTicketRepository) UpdateActiveStatusBySectionID(ctx context.Context, tx *sql.Tx, sectionID int, isActive bool) error {
	query := "UPDATE status_ticket SET is_active = $1, updated_at = NOW() WHERE section_id = $2"
	_, err := tx.ExecContext(ctx, query, isActive, sectionID)
	return err
}

// GET FALLBACK ACTIVE STATUS
func (r *StatusTicketRepository) GetDynamicFallbackStatusID(ctx context.Context, tx *sql.Tx, deactivatedSectionSequence int) (int, error) {
	var fallbackStatusID int
	query := `
        SELECT st.id
        FROM status_ticket st
        JOIN section_status_ticket sst ON st.section_id = sst.id
        WHERE sst.is_active = true
          AND sst.sequence < $1
        ORDER BY sst.sequence DESC, st.sequence DESC
        LIMIT 1`

	err := tx.QueryRowContext(ctx, query, deactivatedSectionSequence).Scan(&fallbackStatusID)
	return fallbackStatusID, err
}

// GET ALL BY SECTION
func (r *StatusTicketRepository) FindAllOrdered() ([]model.StatusTicket, error) {
	query := "SELECT id, name, sequence, is_active, section_id FROM status_ticket ORDER BY section_id, sequence ASC"
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []model.StatusTicket
	for rows.Next() {
		var s model.StatusTicket
		err := rows.Scan(&s.ID, &s.Name, &s.Sequence, &s.IsActive, &s.SectionID)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, s)
	}
	return statuses, nil
}

// UPDATE SEQUENCE
func (r *StatusTicketRepository) UpdateSequence(ctx context.Context, tx *sql.Tx, id int, newSequence int) error {
	query := "UPDATE status_ticket SET sequence = $1, updated_at = NOW() WHERE id = $2"
	_, err := tx.ExecContext(ctx, query, newSequence, id)
	return err
}

// FIND BY SEQUENCE
func (r *StatusTicketRepository) FindBySequence(sequence int) (*model.StatusTicket, error) {
	query := "SELECT id, name, sequence, is_active, created_at, updated_at FROM status_ticket WHERE sequence = $1"
	row := r.DB.QueryRow(query, sequence)

	var s model.StatusTicket
	err := row.Scan(&s.ID, &s.Name, &s.Sequence, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// FIND BY NAME
func (r *StatusTicketRepository) FindByName(name string) (*model.StatusTicket, error) {
	query := "SELECT id, name, sequence, is_active, section_id FROM status_ticket WHERE name = $1"
	row := r.DB.QueryRow(query, name)

	var s model.StatusTicket
	err := row.Scan(&s.ID, &s.Name, &s.Sequence, &s.IsActive, &s.SectionID)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

package repository

import (
	"context"
	"database/sql"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
)

type SectionStatusTicketRepository struct {
	DB *sql.DB
}

func NewSectionStatusTicketRepository(db *sql.DB) *SectionStatusTicketRepository {
	return &SectionStatusTicketRepository{DB: db}
}

// CREATE
func (r *SectionStatusTicketRepository) Create(req dto.CreateSectionStatusTicketRequest) (*model.SectionStatusTicket, error) {
	query := `
        INSERT INTO section_status_ticket (name, sequence, is_active) 
        VALUES ($1, $2, false)
        RETURNING id, name, sequence, is_active, created_at, updated_at`

	row := r.DB.QueryRow(query, req.Name, req.Sequence)

	var newSection model.SectionStatusTicket
	err := row.Scan(
		&newSection.ID, &newSection.Name, &newSection.Sequence, &newSection.IsActive,
		&newSection.CreatedAt, &newSection.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &newSection, nil
}

// GET ALL
func (r *SectionStatusTicketRepository) FindAll() ([]model.SectionStatusTicket, error) {
	query := "SELECT id, name, sequence, is_active, created_at, updated_at FROM section_status_ticket ORDER BY sequence ASC"
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []model.SectionStatusTicket
	for rows.Next() {
		var s model.SectionStatusTicket
		err := rows.Scan(&s.ID, &s.Name, &s.Sequence, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		sections = append(sections, s)
	}
	return sections, nil
}

// GET BY ID
func (r *SectionStatusTicketRepository) FindByID(id int) (*model.SectionStatusTicket, error) {
	query := "SELECT id, name, sequence, is_active, created_at, updated_at FROM section_status_ticket WHERE id = $1"
	row := r.DB.QueryRow(query, id)
	var s model.SectionStatusTicket
	err := row.Scan(&s.ID, &s.Name, &s.Sequence, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

// CHANGE STATUS
func (r *SectionStatusTicketRepository) UpdateActiveStatus(ctx context.Context, tx *sql.Tx, id int, isActive bool) error {
	query := "UPDATE section_status_ticket SET is_active = $1, updated_at = NOW() WHERE id = $2"
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

// UPDATE NAME
func (r *SectionStatusTicketRepository) Update(id int, req dto.UpdateSectionStatusTicketRequest) (*model.SectionStatusTicket, error) {
	query := `UPDATE section_status_ticket SET name = $1, updated_at = NOW() WHERE id = $2
              RETURNING id, name, sequence, is_active, created_at, updated_at`
	row := r.DB.QueryRow(query, req.Name, id)
	var updatedSection model.SectionStatusTicket
	err := row.Scan(&updatedSection.ID, &updatedSection.Name, &updatedSection.Sequence, &updatedSection.IsActive, &updatedSection.CreatedAt, &updatedSection.UpdatedAt)
	return &updatedSection, err
}

// DELETE
func (r *SectionStatusTicketRepository) Delete(id int) error {
	query := "DELETE FROM section_status_ticket WHERE id = $1"
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

// REORDER
func (r *SectionStatusTicketRepository) UpdateSequence(ctx context.Context, tx *sql.Tx, id int, newSequence int) error {
	query := "UPDATE section_status_ticket SET sequence = $1, updated_at = NOW() WHERE id = $2"
	_, err := tx.ExecContext(ctx, query, newSequence, id)
	return err
}

// VALIDATION HELPER
func (r *SectionStatusTicketRepository) CountActiveSections() (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM section_status_ticket WHERE is_active = true"
	err := r.DB.QueryRow(query).Scan(&count)
	return count, err
}

// HELPER UNIQUE NAME
func (r *SectionStatusTicketRepository) IsNameTaken(name string, currentID int) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM section_status_ticket WHERE name = $1 AND id != $2)"
	err := r.DB.QueryRow(query, name, currentID).Scan(&exists)
	return exists, err
}

// HELPER COUNT ALL
func (r *SectionStatusTicketRepository) CountAll() (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM section_status_ticket"
	err := r.DB.QueryRow(query).Scan(&count)
	return count, err
}

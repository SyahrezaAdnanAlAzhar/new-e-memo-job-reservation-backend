package repository

import (
	"context"
	"database/sql"
	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
)

type EmployeePositionRepository struct {
	DB *sql.DB
}

func NewEmployeePositionRepository(db *sql.DB) *EmployeePositionRepository {
	return &EmployeePositionRepository{DB: db}
}

// CREATE
func (r *EmployeePositionRepository) Create(ctx context.Context, tx *sql.Tx, req dto.CreateEmployeePositionRequest) (*model.EmployeePosition, error) {
	query := `
        INSERT INTO employee_position (name, is_active) 
        VALUES ($1, true)
        RETURNING id, name, is_active, created_at, updated_at`

	row := tx.QueryRowContext(ctx, query, req.Name)

	var newPos model.EmployeePosition
	err := row.Scan(&newPos.ID, &newPos.Name, &newPos.IsActive, &newPos.CreatedAt, &newPos.UpdatedAt)
	return &newPos, err
}

// GET ALL
func (r *EmployeePositionRepository) FindAll() ([]model.EmployeePosition, error) {
	query := "SELECT id, name, is_active, created_at, updated_at FROM employee_position ORDER BY id ASC"
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []model.EmployeePosition
	for rows.Next() {
		var p model.EmployeePosition
		err := rows.Scan(&p.ID, &p.Name, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	return positions, nil
}

// GET BY ID
func (r *EmployeePositionRepository) FindByID(id int) (*model.EmployeePosition, error) {
	query := "SELECT id, name, is_active, created_at, updated_at FROM employee_position WHERE id = $1"
	row := r.DB.QueryRow(query, id)

	var p model.EmployeePosition
	err := row.Scan(&p.ID, &p.Name, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	return &p, err
}

// UPDATE
func (r *EmployeePositionRepository) Update(id int, req dto.UpdateEmployeePositionRequest) (*model.EmployeePosition, error) {
	query := `
        UPDATE employee_position SET name = $1, updated_at = NOW() WHERE id = $2
        RETURNING id, name, is_active, created_at, updated_at`
	row := r.DB.QueryRow(query, req.Name, id)
	var updatedPos model.EmployeePosition
	err := row.Scan(&updatedPos.ID, &updatedPos.Name, &updatedPos.IsActive, &updatedPos.CreatedAt, &updatedPos.UpdatedAt)
	return &updatedPos, err
}

// DELETE
func (r *EmployeePositionRepository) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	query := "DELETE FROM employee_position WHERE id = $1"
	result, err := tx.ExecContext(ctx, query, id)
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
func (r *EmployeePositionRepository) UpdateActiveStatus(id int, isActive bool) error {
	query := "UPDATE employee_position SET is_active = $1, updated_at = NOW() WHERE id = $2"
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

// MAKE STATUS TICKET "DIBATALKAN" CAUSE OF IN ACTIVE EMPLOYEE POSITION
func (r *TicketRepository) CancelTicketsByPosition(ctx context.Context, tx *sql.Tx, positionID int, cancelledStatusID int) error {
	finishQuery := `
        UPDATE track_status_ticket
        SET finish_date = NOW()
        WHERE finish_date IS NULL
        AND ticket_id IN (
            SELECT t.id FROM ticket t
            JOIN employee e ON t.requestor = e.npk
            WHERE e.employee_position_id = $1
        )`
	_, err := tx.ExecContext(ctx, finishQuery, positionID)
	if err != nil {
		return err
	}

	createQuery := `
        INSERT INTO track_status_ticket (ticket_id, status_ticket_id, start_date)
        SELECT t.id, $1, NOW()
        FROM ticket t
        JOIN employee e ON t.requestor = e.npk
        WHERE e.employee_position_id = $2`
	_, err = tx.ExecContext(ctx, createQuery, cancelledStatusID, positionID)
	return err
}

// HELPER
func (r *EmployeePositionRepository) IsNameTaken(name string, currentID int) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM employee_position WHERE name = $1 AND id != $2)"
	err := r.DB.QueryRow(query, name, currentID).Scan(&exists)
	return exists, err
}

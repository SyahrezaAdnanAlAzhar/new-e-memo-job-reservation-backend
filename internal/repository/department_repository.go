package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
)

type DepartmentRepository struct {
	DB *sql.DB
}

func NewDepartmentRepository(db *sql.DB) *DepartmentRepository {
	return &DepartmentRepository{DB: db}
}

// HELPER

// CHECK UNIQUE NAME
func (r *DepartmentRepository) IsNameTaken(name string, currentID int) (bool, error) {
	var existsID int
	query := "SELECT id FROM department WHERE name = $1 AND id != $2 LIMIT 1"

	err := r.DB.QueryRow(query, name, currentID).Scan(&existsID)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// MAIN

// CREATE
func (r *DepartmentRepository) Create(req dto.CreateDepartmentRequest) (*model.Department, error) {
	query := `
        INSERT INTO department (name, receive_job, is_active)
        VALUES ($1, $2, false)
        RETURNING id, name, receive_job, is_active, created_at, updated_at`

	row := r.DB.QueryRow(query, req.Name, req.ReceiveJob)

	var newDept model.Department
	err := row.Scan(
		&newDept.ID, &newDept.Name, &newDept.ReceiveJob,
		&newDept.IsActive, &newDept.CreatedAt, &newDept.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &newDept, nil
}

// GET ALL
func (r *DepartmentRepository) FindAll(filters dto.DepartmentFilter) ([]model.Department, error) {
	query := "SELECT id, name, receive_job, is_active, created_at, updated_at FROM department"
	var conditions []string
	var args []interface{}
	argID := 1

	if filters.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argID))
		args = append(args, "%"+filters.Name+"%")
		argID++
	}
	if filters.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argID))
		args = append(args, *filters.IsActive)
		argID++
	}
	if filters.ReceiveJob != nil {
		conditions = append(conditions, fmt.Sprintf("receive_job = $%d", argID))
		args = append(args, *filters.ReceiveJob)
		argID++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY id ASC"

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []model.Department
	for rows.Next() {
		var d model.Department
		err := rows.Scan(&d.ID, &d.Name, &d.ReceiveJob, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}
		departments = append(departments, d)
	}
	return departments, nil
}

// GET BY ID
func (r *DepartmentRepository) FindByID(id int) (*model.Department, error) {
	query := "SELECT id, name, receive_job, is_active, created_at, updated_at FROM department WHERE id = $1"
	row := r.DB.QueryRow(query, id)

	var d model.Department
	err := row.Scan(&d.ID, &d.Name, &d.ReceiveJob, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// GET BY NAME
func (r *DepartmentRepository) FindByName(name string) (*model.Department, error) {
	query := "SELECT id, name, receive_job, is_active, created_at, updated_at FROM department WHERE name ILIKE $1"
	row := r.DB.QueryRow(query, name)

	var d model.Department
	err := row.Scan(&d.ID, &d.Name, &d.ReceiveJob, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// DELETE
func (r *DepartmentRepository) Delete(id int) error {
	query := "DELETE FROM department WHERE id = $1"
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("department not found or already deleted")
	}

	return nil
}

// UPDATE
func (r *DepartmentRepository) Update(id int, req dto.UpdateDepartmentRequest) (*model.Department, error) {
	query := `
		UPDATE department 
        SET name = $1, receive_job = $2, is_active = $3, updated_at = NOW() 
        WHERE id = $4 
        RETURNING id, name, receive_job, is_active, created_at, updated_at`

	row := r.DB.QueryRow(query, req.Name, req.ReceiveJob, req.IsActive, id)

	var updatedDept model.Department
	err := row.Scan(
		&updatedDept.ID, &updatedDept.Name,
		&updatedDept.ReceiveJob, &updatedDept.IsActive,
		&updatedDept.CreatedAt, &updatedDept.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &updatedDept, nil
}

// CHANGE ACTIVE STATUS
func (r *DepartmentRepository) UpdateActiveStatus(id int, isActive bool) error {
	query := "UPDATE department SET is_active = $1, updated_at = NOW() WHERE id = $2"
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

// CHECK IS RECEIVE JOB OR NOT
func (r *DepartmentRepository) IsReceiver(departmentID int) (bool, error) {
	var canReceiveJob bool
	query := "SELECT receive_job FROM department WHERE id = $1 AND is_active = true"

	err := r.DB.QueryRow(query, departmentID).Scan(&canReceiveJob)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("department not found or is not active")
		}
		return false, err
	}

	return canReceiveJob, nil
}

// FindRequestorDepartmentOptions mencari daftar unik departemen requestor
// berdasarkan filter tiket yang relevan.
func (r *DepartmentRepository) FindRequestorDepartmentOptions(filters dto.DepartmentOptionsFilter) ([]dto.DepartmentOptionResponse, error) {
	query := `
        SELECT DISTINCT
            req_dept.id,
            req_dept.name
        FROM ticket t
        JOIN employee req_emp ON t.requestor = req_emp.npk
        JOIN department req_dept ON req_emp.department_id = req_dept.id
        JOIN (
            SELECT DISTINCT ON (ticket_id) ticket_id, status_ticket_id
            FROM track_status_ticket
            ORDER BY ticket_id, start_date DESC, id DESC
        ) current_tst ON t.id = current_tst.ticket_id
        JOIN status_ticket st ON current_tst.status_ticket_id = st.id
        WHERE ($1 = 0 OR st.section_id = $1)
          AND ($2 = 0 OR t.department_target_id = $2)
        ORDER BY req_dept.name ASC`

	rows, err := r.DB.Query(query, filters.SectionID, filters.DepartmentTargetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []dto.DepartmentOptionResponse
	for rows.Next() {
		var d dto.DepartmentOptionResponse
		if err := rows.Scan(&d.ID, &d.Name); err != nil {
			return nil, err
		}
		departments = append(departments, d)
	}
	return departments, nil
}

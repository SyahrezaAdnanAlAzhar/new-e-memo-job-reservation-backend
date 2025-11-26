package repository

import (
	"database/sql"
	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"strconv"
	"strings"
)

type UpdateAreaStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type AreaRepository struct {
	DB *sql.DB
}

func NewAreaRepository(db *sql.DB) *AreaRepository {
	return &AreaRepository{DB: db}
}

// HELPER

// CHECK UNIQUE NAME
func (r *AreaRepository) IsNameTakenInDepartment(name string, departmentID int, currentAreaID int) (bool, error) {
	var existsID int
	query := `
        SELECT id FROM area 
        WHERE name = $1 AND department_id = $2 AND id != $3 
        LIMIT 1`

	err := r.DB.QueryRow(query, name, departmentID, currentAreaID).Scan(&existsID)
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
func (r *AreaRepository) Create(req dto.CreateAreaRequest) (*model.Area, error) {
	query := `
        INSERT INTO area (department_id, name, is_active) 
        VALUES ($1, $2, false)
        RETURNING id, department_id, name, is_active, created_at, updated_at`

	row := r.DB.QueryRow(query, req.DepartmentID, req.Name)

	var newArea model.Area
	err := row.Scan(
		&newArea.ID, &newArea.DepartmentID, &newArea.Name,
		&newArea.IsActive, &newArea.CreatedAt, &newArea.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &newArea, nil
}

// GET ALL
func (r *AreaRepository) FindAll(filters map[string]string) ([]model.Area, error) {
	baseQuery := "SELECT id, department_id, name, is_active, created_at, updated_at FROM area"
	var conditions []string
	var args []interface{}
	argID := 1

	if val, ok := filters["is_active"]; ok {
		conditions = append(conditions, "is_active = $"+strconv.Itoa(argID))
		args = append(args, val)
		argID++
	}
	if val, ok := filters["department_id"]; ok {
		conditions = append(conditions, "department_id = $"+strconv.Itoa(argID))
		args = append(args, val)
		argID++
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY department_id ASC, id ASC"

	rows, err := r.DB.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []model.Area
	for rows.Next() {
		var a model.Area
		err := rows.Scan(&a.ID, &a.DepartmentID, &a.Name, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		areas = append(areas, a)
	}
	return areas, nil
}

// GET BY ID
func (r *AreaRepository) FindByID(id int) (*model.Area, error) {
	query := "SELECT id, department_id, name, is_active, created_at, updated_at FROM area WHERE id = $1"
	row := r.DB.QueryRow(query, id)

	var a model.Area
	err := row.Scan(&a.ID, &a.DepartmentID, &a.Name, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

// DELETE
func (r *AreaRepository) Delete(id int) error {
	query := "DELETE FROM area WHERE id = $1"
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// UPDATE
func (r *AreaRepository) Update(id int, req dto.UpdateAreaRequest) (*model.Area, error) {
	query := `
        UPDATE area 
        SET department_id = $1, name = $2, is_active = $3, updated_at = NOW()
        WHERE id = $4
        RETURNING id, department_id, name, is_active, created_at, updated_at`

	row := r.DB.QueryRow(query, req.DepartmentID, req.Name, req.IsActive, id)

	var updatedArea model.Area
	err := row.Scan(
		&updatedArea.ID, &updatedArea.DepartmentID, &updatedArea.Name,
		&updatedArea.IsActive, &updatedArea.CreatedAt, &updatedArea.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &updatedArea, nil
}

// CHANGE ACTIVE STATUS
func (r *AreaRepository) UpdateActiveStatus(id int, isActive bool) error {
	query := "UPDATE area SET is_active = $1, updated_at = NOW() WHERE id = $2"
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

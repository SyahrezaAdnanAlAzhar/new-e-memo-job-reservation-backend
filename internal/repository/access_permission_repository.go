package repository

import (
	"database/sql"
	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
)

type AccessPermissionRepository struct {
	DB *sql.DB
}

type UpdateAccessPermissionStatusRequest struct {
	IsActive bool `json:"is_active"`
}

func NewAccessPermissionRepository(db *sql.DB) *AccessPermissionRepository {
	return &AccessPermissionRepository{DB: db}
}

// CREATE
func (r *AccessPermissionRepository) Create(req dto.CreateAccessPermissionRequest) (*model.AccessPermission, error) {
	query := `
        INSERT INTO access_permission (name, is_active) 
        VALUES ($1, false)
        RETURNING id, name, is_active, created_at, updated_at`

	row := r.DB.QueryRow(query, req.Name)

	var newPermission model.AccessPermission
	err := row.Scan(
		&newPermission.ID, &newPermission.Name, &newPermission.IsActive,
		&newPermission.CreatedAt, &newPermission.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &newPermission, nil
}

// GET ALL
func (r *AccessPermissionRepository) FindAll() ([]model.AccessPermission, error) {
	query := "SELECT id, name, is_active, created_at, updated_at FROM access_permission ORDER BY id ASC"
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []model.AccessPermission
	for rows.Next() {
		var p model.AccessPermission
		err := rows.Scan(&p.ID, &p.Name, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}
	return permissions, nil
}

// GET BY ID
func (r *AccessPermissionRepository) FindByID(id int) (*model.AccessPermission, error) {
	query := "SELECT id, name, is_active, created_at, updated_at FROM access_permission WHERE id = $1"
	row := r.DB.QueryRow(query, id)

	var p model.AccessPermission
	err := row.Scan(&p.ID, &p.Name, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UPDATE
func (r *AccessPermissionRepository) Update(id int, req dto.UpdateAccessPermissionRequest) (*model.AccessPermission, error) {
	query := `
        UPDATE access_permission 
        SET name = $1, updated_at = NOW()
        WHERE id = $2
        RETURNING id, name, is_active, created_at, updated_at`

	row := r.DB.QueryRow(query, req.Name, id)

	var updatedPermission model.AccessPermission
	err := row.Scan(
		&updatedPermission.ID, &updatedPermission.Name, &updatedPermission.IsActive,
		&updatedPermission.CreatedAt, &updatedPermission.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &updatedPermission, nil
}

// DELETE
func (r *AccessPermissionRepository) Delete(id int) error {
	query := "DELETE FROM access_permission WHERE id = $1"
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
func (r *AccessPermissionRepository) UpdateActiveStatus(id int, isActive bool) error {
	query := "UPDATE access_permission SET is_active = $1, updated_at = NOW() WHERE id = $2"
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

// HELPER
// CHECK UNIQUE NAME
func (r *AccessPermissionRepository) IsNameTaken(name string, currentID int) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM access_permission WHERE name = $1 AND id != $2)"
	err := r.DB.QueryRow(query, name, currentID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

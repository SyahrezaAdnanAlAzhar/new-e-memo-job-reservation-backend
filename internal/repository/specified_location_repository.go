package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
)

type SpecifiedLocationRepository struct {
	DB *sql.DB
}

func NewSpecifiedLocationRepository(db *sql.DB) *SpecifiedLocationRepository {
	return &SpecifiedLocationRepository{DB: db}
}

// CREATE
func (r *SpecifiedLocationRepository) Create(req dto.CreateSpecifiedLocationRequest) (*model.SpecifiedLocation, error) {
	query := `
        INSERT INTO specified_location (physical_location_id, name, is_active) 
        VALUES ($1, $2, true)
        RETURNING id, physical_location_id, name, is_active, created_at, updated_at`

	row := r.DB.QueryRow(query, req.PhysicalLocationID, req.Name)

	var newLoc model.SpecifiedLocation
	err := row.Scan(
		&newLoc.ID, &newLoc.PhysicalLocationID, &newLoc.Name, &newLoc.IsActive,
		&newLoc.CreatedAt, &newLoc.UpdatedAt,
	)
	return &newLoc, err
}

// GET ALL
func (r *SpecifiedLocationRepository) FindAll(filters dto.SpecifiedLocationFilter) ([]model.SpecifiedLocation, error) {
	query := "SELECT id, physical_location_id, name, is_active, created_at, updated_at FROM specified_location"
	var conditions []string
	var args []interface{}
	argID := 1

	if filters.PhysicalLocationID != 0 {
		conditions = append(conditions, fmt.Sprintf("physical_location_id = $%d", argID))
		args = append(args, filters.PhysicalLocationID)
		argID++
	}
	if filters.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argID))
		args = append(args, *filters.IsActive)
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

	var locations []model.SpecifiedLocation
	for rows.Next() {
		var loc model.SpecifiedLocation
		err := rows.Scan(
			&loc.ID, &loc.PhysicalLocationID, &loc.Name, &loc.IsActive,
			&loc.CreatedAt, &loc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}
	return locations, nil
}

// GET ALL BY PHYSICAL LOCATION ID
func (r *SpecifiedLocationRepository) FindByPhysicalLocationID(physicalLocationID int) ([]model.SpecifiedLocation, error) {
	query := "SELECT id, physical_location_id, name, is_active, created_at, updated_at FROM specified_location WHERE physical_location_id = $1 ORDER BY id ASC"
	rows, err := r.DB.Query(query, physicalLocationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []model.SpecifiedLocation
	for rows.Next() {
		var loc model.SpecifiedLocation
		err := rows.Scan(
			&loc.ID, &loc.PhysicalLocationID, &loc.Name, &loc.IsActive,
			&loc.CreatedAt, &loc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}
	return locations, nil
}

// GET BY ID
func (r *SpecifiedLocationRepository) FindByID(id int) (*model.SpecifiedLocation, error) {
	query := "SELECT id, physical_location_id, name, is_active, created_at, updated_at FROM specified_location WHERE id = $1"
	row := r.DB.QueryRow(query, id)
	var loc model.SpecifiedLocation
	err := row.Scan(
		&loc.ID, &loc.PhysicalLocationID, &loc.Name, &loc.IsActive,
		&loc.CreatedAt, &loc.UpdatedAt,
	)
	return &loc, err
}

// UPDATE
func (r *SpecifiedLocationRepository) Update(id int, req dto.UpdateSpecifiedLocationRequest) (*model.SpecifiedLocation, error) {
	query := `
        UPDATE specified_location 
        SET physical_location_id = $1, name = $2, updated_at = NOW()
        WHERE id = $3
        RETURNING id, physical_location_id, name, is_active, created_at, updated_at`

	row := r.DB.QueryRow(query, req.PhysicalLocationID, req.Name, id)

	var updatedLoc model.SpecifiedLocation
	err := row.Scan(
		&updatedLoc.ID, &updatedLoc.PhysicalLocationID, &updatedLoc.Name, &updatedLoc.IsActive,
		&updatedLoc.CreatedAt, &updatedLoc.UpdatedAt,
	)
	return &updatedLoc, err
}

// DELETE
func (r *SpecifiedLocationRepository) Delete(id int) error {
	query := "DELETE FROM specified_location WHERE id = $1"
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
func (r *SpecifiedLocationRepository) UpdateActiveStatus(id int, isActive bool) error {
	query := "UPDATE specified_location SET is_active = $1, updated_at = NOW() WHERE id = $2"
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

// HELPER UNIQUE NAME ON IN PHYSICAL LOCATION
func (r *SpecifiedLocationRepository) IsNameTakenInPhysicalLocation(name string, physicalLocationID int, currentID int) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM specified_location WHERE name = $1 AND physical_location_id = $2 AND id != $3)"
	err := r.DB.QueryRow(query, name, physicalLocationID, currentID).Scan(&exists)
	return exists, err
}

func (r *SpecifiedLocationRepository) FindOrCreate(ctx context.Context, tx *sql.Tx, name string, physicalLocationID int) (int, error) {
	var locationID int

	querySelect := "SELECT id FROM specified_location WHERE name ILIKE $1 AND physical_location_id = $2"
	err := tx.QueryRowContext(ctx, querySelect, name, physicalLocationID).Scan(&locationID)

	if err != nil {
		if err == sql.ErrNoRows {
			queryInsert := `
                INSERT INTO specified_location (name, physical_location_id, is_active) 
                VALUES ($1, $2, true)
                RETURNING id`

			errInsert := tx.QueryRowContext(ctx, queryInsert, name, physicalLocationID).Scan(&locationID)
			if errInsert != nil {
				return 0, errInsert
			}
			return locationID, nil
		}
		return 0, err
	}

	return locationID, nil
}

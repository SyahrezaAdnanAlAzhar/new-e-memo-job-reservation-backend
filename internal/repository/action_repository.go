package repository

import (
	"database/sql"

	"e-memo-job-reservation-api/internal/model"
)

type ActionRepository struct {
	DB *sql.DB
}

func NewActionRepository(db *sql.DB) *ActionRepository {
	return &ActionRepository{DB: db}
}

func (r *ActionRepository) FindAll() ([]model.Action, error) {
	query := "SELECT id, name, is_active, hex_code, created_at, updated_at FROM action ORDER BY id ASC"
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []model.Action
	for rows.Next() {
		var a model.Action
		err := rows.Scan(&a.ID, &a.Name, &a.IsActive, &a.HexCode, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	return actions, nil
}

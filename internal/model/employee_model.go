package model

import (
	"database/sql"
	"time"
)

type Employee struct {
	NPK          string        `json:"npk"`
	DepartmentID int           `json:"department_id"`
	AreaID       sql.NullInt64 `json:"area_id"`
	Name         string        `json:"name"`
	IsActive     bool          `json:"is_active"`
	Position     Position      `json:"position"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type Position struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

package model

import "time"

type Area struct {
	ID           int       `json:"id"`
	DepartmentID int       `json:"department_id"`
	Name         string    `json:"name"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

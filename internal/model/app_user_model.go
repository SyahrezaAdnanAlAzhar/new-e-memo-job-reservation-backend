package model

import (
	"database/sql"
	"time"
)

type AppUser struct {
	ID                 int            `json:"id"`
	Username           string         `json:"username"`
	PasswordHash       string         `json:"-"`
	UserType           string         `json:"user_type"`
	EmployeeNPK        sql.NullString `json:"employee_npk"`
	EmployeePositionID int            `json:"employee_position_id"`
	IsActive           bool           `json:"is_active"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

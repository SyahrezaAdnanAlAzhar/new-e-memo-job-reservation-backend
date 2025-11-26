package model

import "time"

type PositionPermission struct {
	EmployeePositionID int       `json:"employee_position_id"`
	AccessPermissionID int       `json:"access_permission_id"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
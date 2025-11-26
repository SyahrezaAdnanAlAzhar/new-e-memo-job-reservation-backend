package dto

import (
	"database/sql"
)

type EmployeeFilter struct {
	DepartmentID       int    `form:"department_id"`
	AreaID             int    `form:"area_id"`
	EmployeePositionID int    `form:"employee_position_id"`
	Search             string `form:"search"` // <-- DIUBAH
	IsActive           *bool  `form:"is_active"`

	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type EmployeeOptionsFilter struct {
	Role               string `form:"role"`
	SectionID          int    `form:"section_id"`
	DepartmentTargetID int    `form:"department_target_id"`
}

type Pagination struct {
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	TotalItems  int64 `json:"total_items"`
	PageSize    int   `json:"page_size"`
}

type PaginatedEmployeeResponse struct {
	Data []EmployeeDetailResponse `json:"data"`

	Pagination Pagination `json:"pagination"`
}

type EmployeeOptionResponse struct {
	NPK  string `json:"npk"`
	Name string `json:"name"`
}

type CreateEmployeeRequest struct {
	NPK                string `json:"npk" binding:"required"`
	Name               string `json:"name" binding:"required"`
	DepartmentID       int    `json:"department_id" binding:"required,gt=0"`
	AreaID             *int   `json:"area_id"`
	EmployeePositionID int    `json:"employee_position_id" binding:"required,gt=0"`
}

type UpdateEmployeeRequest struct {
	Name               string `json:"name" binding:"required"`
	DepartmentID       int    `json:"department_id" binding:"required,gt=0"`
	AreaID             *int   `json:"area_id"`
	EmployeePositionID int    `json:"employee_position_id" binding:"required,gt=0"`
}

type UpdateEmployeeStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type EmployeeDetailResponse struct {
	NPK            string        `json:"npk"`
	Name           string        `json:"name"`
	IsActive       bool          `json:"is_active"`
	DepartmentID   int           `json:"department_id"`
	DepartmentName string        `json:"department_name"`
	AreaID         sql.NullInt64 `json:"area_id"`
	AreaName       *string       `json:"area_name"`
	Position       Position      `json:"position"`
}

type Position struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

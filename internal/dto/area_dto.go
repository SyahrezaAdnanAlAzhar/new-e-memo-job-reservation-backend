package dto

type CreateAreaRequest struct {
	DepartmentID int    `json:"department_id" binding:"required"`
	Name         string `json:"name" binding:"required"`
}

type UpdateAreaRequest struct {
	DepartmentID int    `json:"department_id" binding:"required,gt=0"`
	Name         string `json:"name" binding:"required"`
	IsActive     bool   `json:"is_active"`
}

package dto

type CreatePositionPermissionRequest struct {
	EmployeePositionID int `json:"employee_position_id" binding:"required,gt=0"`
	AccessPermissionID int `json:"access_permission_id" binding:"required,gt=0"`
}

type UpdatePositionPermissionStatusRequest struct {
	IsActive bool `json:"is_active"`
}
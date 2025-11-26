package dto

type CreateEmployeePositionRequest struct {
	Name       string `json:"name" binding:"required"`
	WorkflowID int    `json:"workflow_id" binding:"required,gt=0"`
}

type UpdateEmployeePositionRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateEmployeePositionStatusRequest struct {
	IsActive bool `json:"is_active"`
}
package dto

type CreateDepartmentRequest struct {
	Name       string `json:"name" binding:"required"`
	ReceiveJob bool   `json:"receive_job"`
}

type UpdateDepartmentRequest struct {
	Name       string `json:"name" binding:"required"`
	ReceiveJob bool   `json:"receive_job"`
	IsActive   bool   `json:"is_active"`
}

type UpdateStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type DepartmentFilter struct {
	Name       string `form:"name"`
	IsActive   *bool  `form:"is_active"`
	ReceiveJob *bool  `form:"receive_job"`
}

type DepartmentOptionsFilter struct {
	SectionID          int `form:"section_id"`
	DepartmentTargetID int `form:"department_target_id"`
}

type DepartmentOptionResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
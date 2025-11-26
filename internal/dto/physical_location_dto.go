package dto

type CreatePhysicalLocationRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdatePhysicalLocationRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdatePhysicalLocationStatusRequest struct {
	IsActive bool `json:"is_active"`
}

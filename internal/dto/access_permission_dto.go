package dto

type CreateAccessPermissionRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateAccessPermissionRequest struct {
	Name string `json:"name" binding:"required"`
}
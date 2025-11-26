package dto

type CreateSpecifiedLocationRequest struct {
	PhysicalLocationID int    `json:"physical_location_id" binding:"required,gt=0"`
	Name               string `json:"name" binding:"required"`
}

type UpdateSpecifiedLocationRequest struct {
	PhysicalLocationID int    `json:"physical_location_id" binding:"required,gt=0"`
	Name               string `json:"name" binding:"required"`
}

type UpdateSpecifiedLocationStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type SpecifiedLocationFilter struct {
	PhysicalLocationID int   `form:"physical_location_id"`
	IsActive           *bool `form:"is_active"`
}
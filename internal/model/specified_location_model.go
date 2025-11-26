package model

import "time"

type SpecifiedLocation struct {
	ID                 int       `json:"id"`
	PhysicalLocationID int       `json:"physical_location_id"`
	Name               string    `json:"name"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
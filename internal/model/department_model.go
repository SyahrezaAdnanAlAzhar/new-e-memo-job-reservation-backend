package model

import "time"

type Department struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	ReceiveJob bool      `json:"receive_job"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
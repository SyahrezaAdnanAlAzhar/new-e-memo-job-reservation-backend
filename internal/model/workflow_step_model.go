package model

import "time"

type WorkflowStep struct {
	ID             int       `json:"id"`
	WorkflowID     int       `json:"workflow_id"`
	StatusTicketID int       `json:"status_ticket_id"`
	StepSequence   int       `json:"step_sequence"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
package model

import "time"

type RejectedTicket struct {
	ID          int64     `json:"id"`
	TicketID    int64     `json:"ticket_id"`
	Rejector    string    `json:"rejector"`
	Feedback    string    `json:"feedback"`
	AlreadySeen bool      `json:"already_seen"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
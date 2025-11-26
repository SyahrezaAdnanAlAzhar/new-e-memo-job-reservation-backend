package model

import "time"

type SectionStatusTicket struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Sequence  int       `json:"sequence"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

package model

import "time"

type StatusTicket struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Sequence  int       `json:"sequence"`
	IsActive  bool      `json:"is_active"`
	SectionID int       `json:"section_id"`
	HexColor string    `json:"hex_color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

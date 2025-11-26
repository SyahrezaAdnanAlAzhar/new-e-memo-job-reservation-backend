package model

import (
	"database/sql"
	"time"
)

type Job struct {
	ID          int            `json:"id"`
	TicketID    int            `json:"ticket_id"`
	PicJob      sql.NullString `json:"pic_job"`
	JobPriority int            `json:"job_priority"`
	ReportFiles []FileMetadata `json:"report_files"`
	Version     int            `json:"version"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Ticket      Ticket         `json:"ticket,omitempty"`
}

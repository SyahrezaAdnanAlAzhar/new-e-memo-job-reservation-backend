package model

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type TicketActionLog struct {
	ID             int64          `json:"id"`
	TicketID       int64          `json:"ticket_id"`
	ActionID       int            `json:"action_id"`
	PerformedByNpk string         `json:"performed_by_npk"`
	DetailsText    sql.NullString `json:"details_text"`
	FilePath       pq.StringArray `json:"file_path"`
	FromStatusID   sql.NullInt32  `json:"from_status_id"`
	ToStatusID     int            `json:"to_status_id"`
	PerformedAt    time.Time      `json:"performed_at"`
}

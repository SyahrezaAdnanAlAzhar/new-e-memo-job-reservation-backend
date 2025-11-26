package model

import "database/sql"

type StatusTransition struct {
	ID             int            `json:"id"`
	FromStatusID   sql.NullInt32  `json:"from_status_id"` 
	ToStatusID     int            `json:"to_status_id"`
	ActionID       int            `json:"action_id"`
	ActorRoleID    int            `json:"actor_role_id"`
	RequiresReason bool           `json:"requires_reason"`
	ReasonLabel    sql.NullString `json:"reason_label"`
	RequiresFile   bool           `json:"requires_file"`
}

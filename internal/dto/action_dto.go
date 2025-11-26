package dto

import "time"

type ActionResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	HexCode   string    `json:"hex_code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AvailableTicketActionResponse struct {
	ActionName    string  `json:"action_name"`
	ActionID      int     `json:"-"`
	ToStatusID    int     `json:"-"`
	HexCode       *string `json:"hex_code"`
	RequireReason bool    `json:"require_reason"`
	ReasonLabel   *string `json:"reason_label"`
	RequireFile   bool    `json:"require_file"`
}

type TransitionDetail struct {
	RequiredActorRole string
	Action            AvailableTicketActionResponse
}

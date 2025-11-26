package dto

type CreateRejectedTicketRequest struct {
	TicketID int64  `json:"ticket_id" binding:"required,gt=0"`
	Feedback string `json:"feedback" binding:"required"`
}

type UpdateFeedbackRequest struct {
	Feedback string `json:"feedback" binding:"required"`
}

type UpdateAlreadySeenRequest struct {
	AlreadySeen bool `json:"already_seen"`
}
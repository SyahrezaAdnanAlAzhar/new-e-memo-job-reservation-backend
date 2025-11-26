package dto

type CreateWorkflowRequest struct {
	Name            string `json:"name" binding:"required"`
	StatusTicketIDs []int  `json:"status_ticket_ids" binding:"required,min=1"`
}

type AddWorkflowStepRequest struct {
	WorkflowID     int    `json:"workflow_id" binding:"required,gt=0"`
	StatusTicketID int    `json:"status_ticket_id" binding:"required,gt=0"`
	Position       string `json:"position" binding:"required,oneof=start end"`
}

type UpdateWorkflowStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type UpdateWorkflowRequest struct {
	Name string `json:"name" binding:"required"`
}
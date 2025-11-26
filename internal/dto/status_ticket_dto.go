package dto

type CreateStatusTicketRequest struct {
	Name     string `json:"name" binding:"required"`
	Sequence int    `json:"sequence"`
}

type UpdateStatusTicketStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type ReorderStatusTicketsRequest struct {
	DeleteSectionOrder   []int `json:"delete_section_order"`
	ApprovalSectionOrder []int `json:"approval_section_order"`
	ActualSectionOrder   []int `json:"actual_section_order"`
}

type StatusTicketFilter struct {
	SectionID int   `form:"section_id"`
	IsActive  *bool `form:"is_active"`
}
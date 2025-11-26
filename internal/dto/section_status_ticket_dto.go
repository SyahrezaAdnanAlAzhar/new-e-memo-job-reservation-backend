package dto

type CreateSectionStatusTicketRequest struct {
	Name     string `json:"name" binding:"required"`
	Sequence int    `json:"sequence" binding:"required"`
}

type UpdateSectionStatusTicketStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type UpdateSectionStatusTicketRequest struct {
	Name string `json:"name" binding:"required"`
}

type ReorderSectionsRequest struct {
	OrderedSectionIDs []int `json:"ordered_section_ids" binding:"required,min=1"`
}
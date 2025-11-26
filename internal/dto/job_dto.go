package dto

import "time"

type AssignPICRequest struct {
	PicJob string `json:"pic_job" binding:"required"`
}

type ReorderJobsRequest struct {
	DepartmentTargetID int              `json:"department_target_id" binding:"required"`
	Items              []ReorderJobItem `json:"items" binding:"required,min=1"`
}

type ReorderJobItem struct {
	JobID   int `json:"job_id" binding:"required"`
	Version int `json:"version" binding:"required"`
}

type JobDetailResponse struct {
	// CORE INFORMATION
	JobID          int    `json:"job_id"`
	TicketID       int    `json:"ticket_id"`
	Description    string `json:"description"`
	JobPriority    int    `json:"job_priority"`
	TicketPriority int    `json:"ticket_priority"`
	SpendingAmount *int64 `json:"spending_amount"`
	Version        int    `json:"version"`

	// DEPARTMENT INFORMATION
	AssignedDepartmentID   int    `json:"assigned_department_id"`
	AssignedDepartmentName string `json:"assigned_department_name"`

	// STATUS INFORMATION
	CurrentStatus        *string `json:"current_status"`
	CurrentStatusHexCode *string `json:"current_status_hex_code"`
	CurrentSectionName   *string `json:"current_section_name"`

	// PEOPLE INFORMATION
	PicName             *string `json:"pic_name"`
	RequestorName       string  `json:"requestor_name"`
	RequestorDepartment *string `json:"requestor_department"`

	// TIME INFORMATION
	TicketAgeDays *int       `json:"ticket_age_days"`
	Deadline      *time.Time `json:"deadline"`
	DaysRemaining *int       `json:"days_remaining"`
}

type JobFilter struct {
	// FILTER BY ID
	SectionID            int    `form:"section_id"`
	StatusID             int    `form:"status_id"`
	AssignedDepartmentID int    `form:"assigned_department_id"`
	PicNPK               string `form:"pic_npk"`
	RequestorNPK         string `form:"requestor_npk"`

	// FILTER BY SEARCH QUERY
	SearchQuery string `form:"search"`

	// FILTER BY NAME
	AssignedDepartmentName string `form:"assigned_department_name"`

	// SORTING OPTION
	SortBy string `form:"sort_by"`
}

type AvailableActionResponse struct {
	Name string `json:"name"`
}

type DeleteJobFilesRequest struct {
	FilePathsToDelete []string `json:"file_paths_to_delete" binding:"required,min=1"`
}

package handler

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/service"
	"e-memo-job-reservation-api/internal/util"
	"e-memo-job-reservation-api/pkg/filehandler"

	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	queryService    *service.TicketQueryService
	commandService  *service.TicketCommandService
	workflowService *service.TicketWorkflowService
	priorityService *service.TicketPriorityService
	actionService   *service.TicketActionService
}

type TicketHandlerConfig struct {
	QueryService    *service.TicketQueryService
	CommandService  *service.TicketCommandService
	WorkflowService *service.TicketWorkflowService
	PriorityService *service.TicketPriorityService
	ActionService   *service.TicketActionService
}

func NewTicketHandler(cfg *TicketHandlerConfig) *TicketHandler {
	return &TicketHandler{
		queryService:    cfg.QueryService,
		commandService:  cfg.CommandService,
		workflowService: cfg.WorkflowService,
		priorityService: cfg.PriorityService,
		actionService:   cfg.ActionService,
	}
}

// POST /tickets
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	var req dto.CreateTicketRequest

	if err := c.ShouldBind(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	requestorNPK := c.GetString("user_npk")
	if requestorNPK == "" {
		util.ErrorResponse(c, http.StatusUnauthorized, "User NPK not found in token", nil)
		return
	}

	var filesMetadata []model.FileMetadata
	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["support_files"]
		if len(files) > 0 {
			savedMetadata, saveErr := filehandler.SaveFiles(c, files)
			if saveErr != nil {
				util.ErrorResponse(c, http.StatusInternalServerError, "Failed to save uploaded files", nil)
				return
			}
			filesMetadata = savedMetadata
		}
	}

	createdTicket, err := h.commandService.CreateTicket(c.Request.Context(), req, requestorNPK, filesMetadata)
	if err != nil {
		for _, metadata := range filesMetadata {
			os.Remove(metadata.FilePath)
		}
		switch err.Error() {
		case "requestor not found", "no workflow defined for this user's position":
			util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to create ticket", err.Error())
		}
		return
	}

	util.SuccessResponse(c, http.StatusCreated, createdTicket)
}

// GET ALL
func (h *TicketHandler) GetAllTickets(c *gin.Context) {
	var filters dto.TicketFilter
	if err := c.ShouldBindQuery(&filters); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	tickets, err := h.queryService.GetAllTickets(filters)
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve tickets", err.Error())
		return
	}

	if tickets == nil {
		util.SuccessResponse(c, http.StatusOK, []dto.TicketDetailResponse{})
		return
	}

	util.SuccessResponse(c, http.StatusOK, tickets)
}

// GET BY ID
func (h *TicketHandler) GetTicketByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	ticket, err := h.queryService.GetTicketByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Ticket not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve ticket", nil)
		return
	}

	util.SuccessResponse(c, http.StatusOK, ticket)
}

// PUT UPDATE
func (h *TicketHandler) UpdateTicket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ticket ID format", nil)
		return
	}
	userNPK := c.GetString("user_npk")
	var req dto.UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err = h.commandService.UpdateTicket(c.Request.Context(), id, req, userNPK)
	if err != nil {
		switch err.Error() {
		case "ticket not found", "user not found", "original requestor not found":
			util.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		case "user is not authorized to edit this ticket":
			util.ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
		case "ticket can only be edited if status is 'Ditolak'":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		case "data conflict: ticket has been modified by another user, please refresh":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		case "invalid deadline format, please use YYYY-MM-DD":
			util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update ticket", err.Error())
		}
		return
	}

	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Ticket data updated successfully. Please use the 'Revisi' action to continue the workflow."})
}

// PUT REORDER
func (h *TicketHandler) ReorderTickets(c *gin.Context) {
	userNPK := c.GetString("user_npk")
	var req dto.ReorderTicketsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	err := h.priorityService.ReorderTickets(c.Request.Context(), req, userNPK)
	if err != nil {
		switch err.Error() {
		case "data conflict: one or more tickets have been modified by another user, please refresh":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		case "action performer not found":
			util.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to reorder tickets", err.Error())
		}
		return
	}
	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Ticket priorities updated successfully"})
}

// POST /tickets/:id/action
func (h *TicketHandler) ExecuteAction(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ticket ID format", nil)
		return
	}
	userNPK := c.GetString("user_npk")

	var req dto.ExecuteActionRequest
	if err := c.ShouldBind(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	var filesMetadata []model.FileMetadata
	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["Files"]
		if len(files) > 0 {
			savedMetadata, saveErr := filehandler.SaveFiles(c, files)
			if saveErr != nil {
				util.ErrorResponse(c, http.StatusInternalServerError, "Failed to save uploaded files", nil)
				return
			}
			filesMetadata = savedMetadata
		}
	}

	err = h.workflowService.ExecuteAction(c.Request.Context(), id, userNPK, req, filesMetadata)
	if err != nil {
		for _, metadata := range filesMetadata {
			os.Remove(metadata.FilePath)
		}
		switch err.Error() {
		case "ticket not found", "user not found", "original requestor not found":
			util.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		case "user does not have the required role for this action":
			util.ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
		case "action not allowed from the current status", "reason is required for this action", "file upload is required for this action":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to execute action", err.Error())
		}
		return
	}

	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Action '" + req.ActionName + "' executed successfully"})
}

// GET /tickets/:id/available-actions
func (h *TicketHandler) GetAvailableActions(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userNPK := c.GetString("user_npk")

	actions, err := h.actionService.GetAvailableActions(c.Request.Context(), id, userNPK)
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to get available actions", err.Error())
		return
	}

	if actions == nil {
		util.SuccessResponse(c, http.StatusOK, []dto.AvailableTicketActionResponse{})
		return
	}
	util.SuccessResponse(c, http.StatusOK, actions)
}

// POST /tickets/:id/files
func (h *TicketHandler) AddSupportFiles(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ticket ID format", nil)
		return
	}
	userNPK := c.GetString("user_npk")

	form, err := c.MultipartForm()
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid form data", err.Error())
		return
	}

	files := form.File["files"]

	if len(files) == 0 {
		util.ErrorResponse(c, http.StatusBadRequest, "At least one file must be uploaded", nil)
		return
	}

	err = h.commandService.AddSupportFiles(c.Request.Context(), c, id, userNPK, files)
	if err != nil {
		switch err.Error() {
		case "ticket not found":
			util.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		case "user is not authorized to add files to this ticket":
			util.ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
		case "failed to save one or more files":
			util.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to add support files", err.Error())
		}
		return
	}

	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Files uploaded and added to ticket successfully"})
}

// DELETE /tickets/:id/files
func (h *TicketHandler) RemoveSupportFiles(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ticket ID format", nil)
		return
	}
	userNPK := c.GetString("user_npk")

	var req dto.DeleteFilesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err = h.commandService.RemoveSupportFiles(c.Request.Context(), id, userNPK, req)
	if err != nil {
		switch err.Error() {
		case "ticket not found":
			util.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		case "user is not authorized to remove files from this ticket":
			util.ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove support files", err.Error())
		}
		return
	}

	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Selected files removed successfully"})
}

// GET /reports/ticket-summary
func (h *TicketHandler) GetTicketSummary(c *gin.Context) {
	var filters dto.TicketSummaryFilter
	if err := c.ShouldBindQuery(&filters); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	summary, err := h.queryService.GetTicketSummary(filters)
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve ticket summary", err.Error())
		return
	}

	if summary == nil {
		util.SuccessResponse(c, http.StatusOK, []dto.TicketSummaryResponse{})
		return
	}

	util.SuccessResponse(c, http.StatusOK, summary)
}

// GET /reports/oldest-ticket
func (h *TicketHandler) GetOldestTicket(c *gin.Context) {
	oldestTicket, err := h.queryService.GetOldestTicket()
	if err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "No tickets found in the system", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve oldest ticket", err.Error())
		return
	}

	util.SuccessResponse(c, http.StatusOK, oldestTicket)
}

// GET /tickets/:id/last-rejection
func (h *TicketHandler) GetLastRejectionDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ticket ID format", nil)
		return
	}

	rejectionDetail, err := h.queryService.GetLastRejectionDetail(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			util.SuccessResponse(c, http.StatusOK, nil)
			return
		}
		if err.Error() == "ticket not found" {
			util.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve last rejection detail", err.Error())
		return
	}

	util.SuccessResponse(c, http.StatusOK, rejectionDetail)
}

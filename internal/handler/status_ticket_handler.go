package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/service"
	"e-memo-job-reservation-api/internal/util"

	"github.com/gin-gonic/gin"
)

type StatusTicketHandler struct {
	service *service.StatusTicketService
}

func NewStatusTicketHandler(service *service.StatusTicketService) *StatusTicketHandler {
	return &StatusTicketHandler{service: service}
}

// POST /status-ticket
func (h *StatusTicketHandler) CreateStatusTicket(c *gin.Context) {
	var req dto.CreateStatusTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	newStatus, err := h.service.CreateStatusTicket(req)
	if err != nil {
		if err.Error() == "status ticket name or sequence already exists" {
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to create status ticket", nil)
		return
	}

	util.SuccessResponse(c, http.StatusCreated, newStatus)
}

// GET /status-ticket
func (h *StatusTicketHandler) GetAllStatusTickets(c *gin.Context) {
	var filters dto.StatusTicketFilter
	if err := c.ShouldBindQuery(&filters); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	statuses, err := h.service.GetAllStatusTickets(filters)
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve status tickets", err.Error())
		return
	}

	if statuses == nil {
		util.SuccessResponse(c, http.StatusOK, []model.StatusTicket{})
		return
	}
	util.SuccessResponse(c, http.StatusOK, statuses)
}

// GET /status-ticket/:id
func (h *StatusTicketHandler) GetStatusTicketByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid status ticket ID format", nil)
		return
	}

	status, err := h.service.GetStatusTicketByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Status ticket not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve status ticket", nil)
		return
	}

	util.SuccessResponse(c, http.StatusOK, status)
}

// DELETE /status-ticket/:id
func (h *StatusTicketHandler) DeleteStatusTicket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid status ticket ID format", nil)
		return
	}
	if err := h.service.DeleteStatusTicket(id); err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Status ticket not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete status ticket", nil)
		return
	}
	c.Status(http.StatusNoContent) // No content, no response body
}

// PATCH /status-ticket/:id/status
func (h *StatusTicketHandler) UpdateStatusTicketActiveStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid status ticket ID format", nil)
		return
	}

	var req dto.UpdateStatusTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.service.UpdateStatusTicketActiveStatus(id, req); err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Status ticket not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update status ticket", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Status ticket status updated successfully"})
}

// POST /status-ticket/reorder
func (h *StatusTicketHandler) ReorderStatusTickets(c *gin.Context) {
	var req dto.ReorderStatusTicketsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.service.ReorderStatusTickets(req); err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to reorder status tickets", err.Error())
		return
	}
	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Status tickets reordered successfully"})
}

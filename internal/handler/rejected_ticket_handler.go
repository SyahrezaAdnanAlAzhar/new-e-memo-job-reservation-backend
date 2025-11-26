package handler

import (
	"net/http"
	"strconv"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/service"
	"e-memo-job-reservation-api/internal/util"

	"github.com/gin-gonic/gin"
)

type RejectedTicketHandler struct {
	service *service.RejectedTicketService
}

func NewRejectedTicketHandler(service *service.RejectedTicketService) *RejectedTicketHandler {
	return &RejectedTicketHandler{service: service}
}

// POST /rejected-tickets/
func (h *RejectedTicketHandler) CreateRejectedTicket(c *gin.Context) {
	var req dto.CreateRejectedTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	userNPK := c.GetString("user_npk")

	newRejection, err := h.service.CreateRejectedTicket(c.Request.Context(), req, userNPK)
	if err != nil {
		switch err.Error() {
		case "ticket already has an active rejection that has not been seen",
			"ticket is still in 'Ditolak' status from a previous rejection":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to create rejection", err.Error())
		}
		return
	}
	util.SuccessResponse(c, http.StatusCreated, newRejection)
}

// PUT /rejected-tickets/:id/feedback
func (h *RejectedTicketHandler) UpdateFeedback(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}
	userNPK := c.GetString("user_npk")

	var req dto.UpdateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	updatedRejection, err := h.service.UpdateFeedback(c.Request.Context(), id, req, userNPK)
	if err != nil {
		switch err.Error() {
		case "user is not authorized to update this feedback":
			util.ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
		case "rejection record not found", "associated ticket not found":
			util.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update feedback", err.Error())
		}
		return
	}
	util.SuccessResponse(c, http.StatusOK, updatedRejection)
}

// PATCH /rejected-ticket/:id/seen
func (h *RejectedTicketHandler) UpdateAlreadySeen(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}
	userNPK := c.GetString("user_npk")

	var req dto.UpdateAlreadySeenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err = h.service.UpdateAlreadySeen(c.Request.Context(), id, req, userNPK)
	if err != nil {
		switch err.Error() {
		case "user is not authorized to perform this action":
			util.ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
		case "rejection record not found", "associated ticket not found":
			util.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update 'already_seen' status", err.Error())
		}
		return
	}
	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "'already_seen' status updated successfully"})
}

// DELETE /rejected-ticket/:id
func (h *RejectedTicketHandler) DeleteRejectedTicket(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}
	userNPK := c.GetString("user_npk")

	err = h.service.DeleteRejectedTicket(c.Request.Context(), id, userNPK)
	if err != nil {
		switch err.Error() {
		case "user is not authorized to delete this rejection record":
			util.ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
		case "can only delete rejection record if ticket status is 'Ditolak'":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		case "rejection record not found":
			util.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete rejection record", err.Error())
		}
		return
	}
	c.Status(http.StatusNoContent)
}

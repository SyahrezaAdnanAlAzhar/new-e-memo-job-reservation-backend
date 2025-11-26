package handler

import (
	"net/http"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/service"
	"e-memo-job-reservation-api/internal/util"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	service *service.SystemService
}

func NewSystemHandler(service *service.SystemService) *SystemHandler {
	return &SystemHandler{service: service}
}

func (h *SystemHandler) UpdateEditMode(c *gin.Context) {
	var req dto.UpdateEditModeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.service.UpdateEditMode(c.Request.Context(), req)
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update system edit mode", err.Error())
		return
	}

	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "System edit mode updated successfully"})
}

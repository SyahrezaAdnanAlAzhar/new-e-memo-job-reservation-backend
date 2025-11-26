package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/service"
	"e-memo-job-reservation-api/internal/util"

	"github.com/gin-gonic/gin"
)

type PositionPermissionHandler struct {
	service *service.PositionPermissionService
}

func NewPositionPermissionHandler(service *service.PositionPermissionService) *PositionPermissionHandler {
	return &PositionPermissionHandler{service: service}
}

// POST /position-permissions
func (h *PositionPermissionHandler) CreatePositionPermission(c *gin.Context) {
	var req dto.CreatePositionPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	newPerm, err := h.service.CreatePositionPermission(req)
	if err != nil {
		switch err.Error() {
		case "invalid employee_position_id or access_permission_id":
			util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		case "this permission is already assigned to the position":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to create position permission", nil)
		}
		return
	}
	util.SuccessResponse(c, http.StatusCreated, newPerm)
}

// GET /position-permissions
func (h *PositionPermissionHandler) GetAllPositionPermissions(c *gin.Context) {
	permissions, err := h.service.GetAllPositionPermissions()
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve position permissions", nil)
		return
	}

	if permissions == nil {
		util.SuccessResponse(c, http.StatusOK, []gin.H{})
		return
	}
	util.SuccessResponse(c, http.StatusOK, permissions)
}

// PATCH /position-permissions/positions/:posId/permissions/:permId/status
func (h *PositionPermissionHandler) UpdatePositionPermissionActiveStatus(c *gin.Context) {
	posID, err := strconv.Atoi(c.Param("posId"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid position ID format", nil)
		return
	}
	permID, err := strconv.Atoi(c.Param("permId"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID format", nil)
		return
	}

	var req dto.UpdatePositionPermissionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.service.UpdatePositionPermissionActiveStatus(posID, permID, req); err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Position permission not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update status", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Position permission status updated successfully"})
}

// DELETE /api/v1/position-permissions/positions/:posId/permissions/:permId
func (h *PositionPermissionHandler) DeletePositionPermission(c *gin.Context) {
	posID, err := strconv.Atoi(c.Param("posId"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid position ID format", nil)
		return
	}
	permID, err := strconv.Atoi(c.Param("permId"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid permission ID format", nil)
		return
	}

	if err := h.service.DeletePositionPermission(posID, permID); err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Position permission not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete position permission", nil)
		return
	}
	c.Status(http.StatusNoContent)
}

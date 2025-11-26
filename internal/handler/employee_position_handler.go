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

type EmployeePositionHandler struct {
	service *service.EmployeePositionService
}

func NewEmployeePositionHandler(service *service.EmployeePositionService) *EmployeePositionHandler {
	return &EmployeePositionHandler{service: service}
}

// POST /employee-position
func (h *EmployeePositionHandler) CreateEmployeePosition(c *gin.Context) {
	var req dto.CreateEmployeePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	newPos, err := h.service.CreateEmployeePosition(c.Request.Context(), req)
	if err != nil {
		switch err.Error() {
		case "position name already exists":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		case "invalid workflow_id":
			util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to create position", nil)
		return
	}
	util.SuccessResponse(c, http.StatusCreated, newPos)
}

// GET /api/v1/employee-position
func (h *EmployeePositionHandler) GetAllEmployeePositions(c *gin.Context) {
	positions, err := h.service.GetAllEmployeePositions()
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve positions", nil)
		return
	}
	if positions == nil {
		util.SuccessResponse(c, http.StatusOK, []gin.H{})
		return
	}
	util.SuccessResponse(c, http.StatusOK, positions)
}

// GET /employee-position/:id
func (h *EmployeePositionHandler) GetEmployeePositionByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	position, err := h.service.GetEmployeePositionByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Position not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve position", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, position)
}

// PUT /employee-position/:id
func (h *EmployeePositionHandler) UpdateEmployeePosition(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var req dto.UpdateEmployeePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	updatedPos, err := h.service.UpdateEmployeePosition(id, req)
	if err != nil {
		switch {
		case err.Error() == "position name already exists":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		case err == sql.ErrNoRows:
			util.ErrorResponse(c, http.StatusNotFound, "Position not found", nil)
			return
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update position", nil)
			return
		}
	}
	util.SuccessResponse(c, http.StatusOK, updatedPos)
}

// DELETE /employee-position/:id
func (h *EmployeePositionHandler) DeleteEmployeePosition(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	if err := h.service.DeleteEmployeePosition(c.Request.Context(), id); err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Position not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete position", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// PATCH /employee-position/:id/status
func (h *EmployeePositionHandler) UpdateEmployeePositionActiveStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var req dto.UpdateEmployeePositionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.service.UpdateEmployeePositionActiveStatus(id, req); err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Position not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update status", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Position status updated successfully"})
}

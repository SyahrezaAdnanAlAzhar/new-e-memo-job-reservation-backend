package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"
	"e-memo-job-reservation-api/internal/service"
	"e-memo-job-reservation-api/internal/util"

	"github.com/gin-gonic/gin"
)

type AreaHandler struct {
	service *service.AreaService
}

func NewAreaHandler(service *service.AreaService) *AreaHandler {
	return &AreaHandler{service: service}
}

// POST /area
func (h *AreaHandler) CreateArea(c *gin.Context) {
	var req dto.CreateAreaRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	newArea, err := h.service.CreateArea(req)
	if err != nil {
		if err.Error() == "area name already exists in this department" {
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to create area", nil)
		return
	}

	util.SuccessResponse(c, http.StatusCreated, newArea)
}

// GET /area
func (h *AreaHandler) GetAllAreas(c *gin.Context) {
	filters := make(map[string]string)

	if isActive, exists := c.GetQuery("is_active"); exists {
		filters["is_active"] = isActive
	}

	if deptID, exists := c.GetQuery("department_id"); exists {
		filters["department_id"] = deptID
	}

	areas, err := h.service.GetAllAreas(filters)
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve areas", nil)
		return
	}

	if areas == nil {
		util.SuccessResponse(c, http.StatusOK, []model.Area{})
		return
	}

	util.SuccessResponse(c, http.StatusOK, areas)
}

// GET /area/:id
func (h *AreaHandler) GetAreaByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid area ID format", nil)
		return
	}

	area, err := h.service.GetAreaByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Area not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve area", nil)
		return
	}

	util.SuccessResponse(c, http.StatusOK, area)
}

// DELETE /area/:id
func (h *AreaHandler) DeleteArea(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid area ID format", nil)
		return
	}

	err = h.service.DeleteArea(id)
	if err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Area not found or already deleted", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete area", nil)
		return
	}

	c.Status(http.StatusNoContent)
}

// PUT /area/:id
func (h *AreaHandler) UpdateArea(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid area ID format", nil)
		return
	}

	var req dto.UpdateAreaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	updatedArea, err := h.service.UpdateArea(id, req)
	if err != nil {
		if err.Error() == "area name already exists in this department" {
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Area not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update area", nil)
		return
	}

	util.SuccessResponse(c, http.StatusOK, updatedArea)
}

// PATCH /area/:id/status
func (h *AreaHandler) UpdateAreaActiveStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid area ID format", nil)
		return
	}

	var req repository.UpdateAreaStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err = h.service.UpdateAreaActiveStatus(id, req)
	if err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Area not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update area status", nil)
		return
	}

	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Area status updated successfully"})
}

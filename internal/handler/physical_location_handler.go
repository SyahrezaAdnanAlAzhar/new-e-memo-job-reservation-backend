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

type PhysicalLocationHandler struct {
	service *service.PhysicalLocationService
}

func NewPhysicalLocationHandler(service *service.PhysicalLocationService) *PhysicalLocationHandler {
	return &PhysicalLocationHandler{service: service}
}

// POST /physical-location
func (h *PhysicalLocationHandler) CreatePhysicalLocation(c *gin.Context) {
	var req dto.CreatePhysicalLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	newLoc, err := h.service.CreatePhysicalLocation(req)
	if err != nil {
		if err.Error() == "physical location name already exists" {
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to create physical location", nil)
		return
	}
	util.SuccessResponse(c, http.StatusCreated, newLoc)
}

// GET /physical-location
func (h *PhysicalLocationHandler) GetAllPhysicalLocations(c *gin.Context) {
	filters := make(map[string]string)
	if isActive, exists := c.GetQuery("is_active"); exists {
		filters["is_active"] = isActive
	}

	locations, err := h.service.GetAllPhysicalLocations(filters)
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve physical locations", nil)
		return
	}

	if locations == nil {
		util.SuccessResponse(c, http.StatusOK, []model.PhysicalLocation{})
		return
	}
	util.SuccessResponse(c, http.StatusOK, locations)
}

// GET /physical-location/:id
func (h *PhysicalLocationHandler) GetPhysicalLocationByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	location, err := h.service.GetPhysicalLocationByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Physical location not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve physical location", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, location)
}

// PUT /physical-location/:id
func (h *PhysicalLocationHandler) UpdatePhysicalLocation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var req dto.UpdatePhysicalLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	updatedLoc, err := h.service.UpdatePhysicalLocation(id, req)
	if err != nil {
		if err.Error() == "physical location name already exists" {
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Physical location not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update physical location", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, updatedLoc)
}

// DELETE /physical-location/:id
func (h *PhysicalLocationHandler) DeletePhysicalLocation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	if err := h.service.DeletePhysicalLocation(id); err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Physical location not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete physical location", nil)
		return
	}
	c.Status(http.StatusNoContent)
}

// PATCH /physical-location/:id/status
func (h *PhysicalLocationHandler) UpdatePhysicalLocationActiveStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var req dto.UpdatePhysicalLocationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.service.UpdatePhysicalLocationActiveStatus(id, req); err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Physical location not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update status", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Physical location status updated successfully"})
}

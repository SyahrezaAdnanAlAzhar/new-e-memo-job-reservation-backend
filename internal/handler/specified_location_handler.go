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

type SpecifiedLocationHandler struct {
	service *service.SpecifiedLocationService
}

func NewSpecifiedLocationHandler(service *service.SpecifiedLocationService) *SpecifiedLocationHandler {
	return &SpecifiedLocationHandler{service: service}
}

// POST /specified-location
func (h *SpecifiedLocationHandler) CreateSpecifiedLocation(c *gin.Context) {
	var req dto.CreateSpecifiedLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	newLoc, err := h.service.CreateSpecifiedLocation(req)
	if err != nil {
		switch err.Error() {
		case "invalid physical_location_id":
			util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		case "location name already exists in this physical location":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to create specified location", nil)
		}
		return
	}
	util.SuccessResponse(c, http.StatusCreated, newLoc)
}

// GET /specified-location
func (h *SpecifiedLocationHandler) GetAllSpecifiedLocations(c *gin.Context) {
	var filters dto.SpecifiedLocationFilter
	if err := c.ShouldBindQuery(&filters); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	locations, err := h.service.GetAllSpecifiedLocations(filters)
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve specified locations", err.Error())
		return
	}

	if locations == nil {
		util.SuccessResponse(c, http.StatusOK, []model.SpecifiedLocation{})
		return
	}
	util.SuccessResponse(c, http.StatusOK, locations)
}

// GET /specified-location/:id
func (h *SpecifiedLocationHandler) GetSpecifiedLocationByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	location, err := h.service.GetSpecifiedLocationByID(id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			util.ErrorResponse(c, http.StatusNotFound, "Specified location not found", nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve specified location", nil)
		}
		return
	}
	util.SuccessResponse(c, http.StatusOK, location)
}

// PUT /specified-location/:id
func (h *SpecifiedLocationHandler) UpdateSpecifiedLocation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var req dto.UpdateSpecifiedLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	updatedLoc, err := h.service.UpdateSpecifiedLocation(id, req)
	if err != nil {
		switch err.Error() {
		case "location name already exists in this physical location":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		case sql.ErrNoRows.Error():
			util.ErrorResponse(c, http.StatusNotFound, "Specified location not found", nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update specified location", nil)
		}
		return
	}
	util.SuccessResponse(c, http.StatusOK, updatedLoc)
}

// DELETE /specified-location/:id
func (h *SpecifiedLocationHandler) DeleteSpecifiedLocation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	if err := h.service.DeleteSpecifiedLocation(id); err != nil {
		switch err {
		case sql.ErrNoRows:
			util.ErrorResponse(c, http.StatusNotFound, "Specified location not found", nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete specified location", nil)
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// PATCH /specified-location/:id/status
func (h *SpecifiedLocationHandler) UpdateSpecifiedLocationActiveStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var req dto.UpdateSpecifiedLocationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.service.UpdateSpecifiedLocationActiveStatus(id, req); err != nil {
		switch err {
		case sql.ErrNoRows:
			util.ErrorResponse(c, http.StatusNotFound, "Specified location not found", nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update status", nil)
		}
		return
	}
	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Specified location status updated successfully"})
}

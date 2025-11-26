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

type SectionStatusTicketHandler struct {
	service *service.SectionStatusTicketService
}

func NewSectionStatusTicketHandler(service *service.SectionStatusTicketService) *SectionStatusTicketHandler {
	return &SectionStatusTicketHandler{service: service}
}

// POST /section-status-ticket
func (h *SectionStatusTicketHandler) CreateSectionStatusTicket(c *gin.Context) {
	var req dto.CreateSectionStatusTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	newSection, err := h.service.CreateSectionStatusTicket(req)
	if err != nil {
		switch err.Error() {
		case "section name or sequence already exists":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to create section status ticket", nil)
		}
		return
	}
	util.SuccessResponse(c, http.StatusCreated, newSection)
}

// GET /section-status-ticket
func (h *SectionStatusTicketHandler) GetAllSectionStatusTickets(c *gin.Context) {
	sections, err := h.service.GetAllSectionStatusTickets()
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve sections", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, sections)
}

// GET /section-status-ticket/:id
func (h *SectionStatusTicketHandler) GetSectionStatusTicketByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}
	section, err := h.service.GetSectionStatusTicketByID(id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			util.ErrorResponse(c, http.StatusNotFound, "Section not found", nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve section", nil)
		}
		return
	}
	util.SuccessResponse(c, http.StatusOK, section)
}

// PATCH /section-status-ticket/:id/status
func (h *SectionStatusTicketHandler) UpdateSectionStatusTicketActiveStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}
	var req dto.UpdateSectionStatusTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err = h.service.UpdateSectionStatusTicketActiveStatus(c.Request.Context(), id, req)
	if err != nil {
		switch err.Error() {
		case "cannot deactivate, must have at least two active sections",
			"cannot deactivate the first section":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		case sql.ErrNoRows.Error():
			util.ErrorResponse(c, http.StatusNotFound, "Section not found", nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update section status", err.Error())
		}
		return
	}
	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Section status and related data updated successfully"})
}

// PUT /section-status-ticket/:id
func (h *SectionStatusTicketHandler) UpdateSectionStatusTicket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var req dto.UpdateSectionStatusTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	updatedSection, err := h.service.UpdateSectionStatusTicketName(id, req)
	if err != nil {
		switch err.Error() {
		case "section name already exists":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		case sql.ErrNoRows.Error():
			util.ErrorResponse(c, http.StatusNotFound, "Section not found", nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update section", nil)
		}
		return
	}
	util.SuccessResponse(c, http.StatusOK, updatedSection)
}

// DELETE /section-status-ticket/:id
func (h *SectionStatusTicketHandler) DeleteSectionStatusTicket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = h.service.DeleteSectionStatusTicket(id)
	if err != nil {
		switch err.Error() {
		case "cannot delete, must have at least two sections",
			"cannot delete the first section":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		case sql.ErrNoRows.Error():
			util.ErrorResponse(c, http.StatusNotFound, "Section not found", nil)
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete section", nil)
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// PUT /section-status-ticket/reorder
func (h *SectionStatusTicketHandler) ReorderSections(c *gin.Context) {
	var req dto.ReorderSectionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := h.service.ReorderSections(c.Request.Context(), req); err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to reorder sections", err.Error())
		return
	}
	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Sections reordered successfully"})
}

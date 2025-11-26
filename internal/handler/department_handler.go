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

type DepartmentHandler struct {
	service *service.DepartmentService
}

func NewDepartmentHandler(service *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: service}
}

// POST /department
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req dto.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	newDept, err := h.service.CreateDepartment(req)
	if err != nil {
		if err.Error() == "department name already exists" {
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to create department", nil)
		return
	}

	util.SuccessResponse(c, http.StatusCreated, newDept)
}

// GET /department
func (h *DepartmentHandler) GetAllDepartments(c *gin.Context) {
	var filters dto.DepartmentFilter
	if err := c.ShouldBindQuery(&filters); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	departments, err := h.service.GetAllDepartments(filters)
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve departments", err.Error())
		return
	}

	if departments == nil {
		util.SuccessResponse(c, http.StatusOK, []model.Department{})
		return
	}

	util.SuccessResponse(c, http.StatusOK, departments)
}

// GET /department/:id
func (h *DepartmentHandler) GetDepartmentByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid department ID", nil)
		return
	}

	department, err := h.service.GetDepartmentByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Department not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve department", nil)
		return
	}

	util.SuccessResponse(c, http.StatusOK, department)
}

// DELETE /department/:id
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid department ID", nil)
		return
	}

	err = h.service.DeleteDepartment(id)
	if err != nil {
		if err.Error() == "department not found or already deleted" {
			util.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete department", nil)
		return
	}

	c.Status(http.StatusNoContent)
}

// PUT /department/:id
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid department ID", nil)
		return
	}

	var req dto.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	updatedDept, err := h.service.UpdateDepartment(id, req)
	if err != nil {
		if err.Error() == "department name already exists" {
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Department not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update department", nil)
		return
	}

	util.SuccessResponse(c, http.StatusOK, updatedDept)
}

// PATCH /department/:id/status
func (h *DepartmentHandler) UpdateDepartmentActiveStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid department ID", nil)
		return
	}

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err = h.service.UpdateDepartmentActiveStatus(id, req)
	if err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Department not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update department status", nil)
		return
	}

	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Department status updated successfully"})
}

// GET /department/options
func (h *DepartmentHandler) GetRequestorDepartmentOptions(c *gin.Context) {
	var filters dto.DepartmentOptionsFilter
	if err := c.ShouldBindQuery(&filters); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	options, err := h.service.GetRequestorDepartmentOptions(filters)
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve department options", err.Error())
		return
	}

	if options == nil {
		util.SuccessResponse(c, http.StatusOK, []dto.DepartmentOptionResponse{})
		return
	}

	util.SuccessResponse(c, http.StatusOK, options)
}

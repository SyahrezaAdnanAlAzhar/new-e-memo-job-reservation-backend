package handler

import (
	"net/http"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/service"
	"e-memo-job-reservation-api/internal/util"

	"github.com/gin-gonic/gin"
)

type EmployeeHandler struct {
	service *service.EmployeeService
}

func NewEmployeeHandler(service *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: service}
}

func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
	var req dto.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	newEmployee, err := h.service.CreateEmployee(req)
	if err != nil {
		if err.Error() == "employee with this NPK already exists" {
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		if err.Error() == "invalid department_id, area_id, or employee_position_id" {
			util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to create employee", err.Error())
		return
	}
	util.SuccessResponse(c, http.StatusCreated, newEmployee)
}

func (h *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	npk := c.Param("npk")
	var req dto.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	updatedEmployee, err := h.service.UpdateEmployee(npk, req)
	if err != nil {
		if err.Error() == "employee not found" {
			util.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		if err.Error() == "invalid department_id, area_id, or employee_position_id" {
			util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update employee", err.Error())
		return
	}
	util.SuccessResponse(c, http.StatusOK, updatedEmployee)
}

func (h *EmployeeHandler) UpdateEmployeeActiveStatus(c *gin.Context) {
	npk := c.Param("npk")
	var req dto.UpdateEmployeeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.service.UpdateEmployeeActiveStatus(npk, req)
	if err != nil {
		if err.Error() == "employee not found" {
			util.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update employee status", err.Error())
		return
	}
	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Employee status updated successfully"})
}

func (h *EmployeeHandler) GetAllEmployees(c *gin.Context) {
	var filters dto.EmployeeFilter
	if err := c.ShouldBindQuery(&filters); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	paginatedResponse, err := h.service.GetAllEmployees(filters)
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve employees", err.Error())
		return
	}

	util.SuccessResponse(c, http.StatusOK, paginatedResponse)
}

// GET /employee/options
func (h *EmployeeHandler) GetEmployeeOptions(c *gin.Context) {
	var filters dto.EmployeeOptionsFilter
	if err := c.ShouldBindQuery(&filters); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	if filters.Role == "" {
		util.ErrorResponse(c, http.StatusBadRequest, "Query parameter 'role' is required", nil)
		return
	}

	options, err := h.service.GetEmployeeOptions(filters)
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve employee options", err.Error())
		return
	}

	if options == nil {
		util.SuccessResponse(c, http.StatusOK, []dto.EmployeeOptionResponse{})
		return
	}

	util.SuccessResponse(c, http.StatusOK, options)
}

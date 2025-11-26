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

type AccessPermissionHandler struct {
	service *service.AccessPermissionService
}

func NewAccessPermissionHandler(service *service.AccessPermissionService) *AccessPermissionHandler {
	return &AccessPermissionHandler{service: service}
}

// POST /access-permissions
func (h *AccessPermissionHandler) CreateAccessPermission(c *gin.Context) {
	var req dto.CreateAccessPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	newPermission, err := h.service.CreateAccessPermission(req)
	if err != nil {
		if err.Error() == "permission name already exists" {
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to create permission", nil)
		return
	}

	util.SuccessResponse(c, http.StatusCreated, newPermission)
}

// GET /access-permissions
func (h *AccessPermissionHandler) GetAllAccessPermissions(c *gin.Context) {
	permissions, err := h.service.GetAllAccessPermissions()
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve permissions", nil)
		return
	}
	if permissions == nil {
		util.SuccessResponse(c, http.StatusOK, []model.AccessPermission{})
		return
	}
	util.SuccessResponse(c, http.StatusOK, permissions)
}

// GET /access-permissions/:id
func (h *AccessPermissionHandler) GetAccessPermissionByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	permission, err := h.service.GetAccessPermissionByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Permission not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve permission", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, permission)
}

// PUT /access-permissions/:id
func (h *AccessPermissionHandler) UpdateAccessPermission(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var req dto.UpdateAccessPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	updatedPermission, err := h.service.UpdateAccessPermission(id, req)
	if err != nil {
		if err.Error() == "permission name already exists" {
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Permission not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update permission", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, updatedPermission)
}

// DELETE /access-permissions/:id
func (h *AccessPermissionHandler) DeleteAccessPermission(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	if err := h.service.DeleteAccessPermission(id); err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Permission not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete permission", nil)
		return
	}
	c.Status(http.StatusNoContent)
}

// PATCH /access-permissions/:id/status
func (h *AccessPermissionHandler) UpdateAccessPermissionActiveStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var req repository.UpdateAccessPermissionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.service.UpdateAccessPermissionActiveStatus(id, req); err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Permission not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update status", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Permission status updated successfully"})
}

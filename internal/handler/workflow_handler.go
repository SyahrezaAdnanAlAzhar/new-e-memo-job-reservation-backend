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

type WorkflowHandler struct {
	service *service.WorkflowService
}

func NewWorkflowHandler(service *service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{service: service}
}

// POST /workflow
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	var req dto.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	newWorkflow, err := h.service.CreateWorkflowWithSteps(c.Request.Context(), req)
	if err != nil {
		switch err.Error() {
		case "workflow name already exists", "cannot add the same status twice to a workflow":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		case "one or more status_ticket_ids are invalid":
			util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
			return
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to create workflow", nil)
			return
		}
	}
	util.SuccessResponse(c, http.StatusCreated, newWorkflow)
}

// POST /workflow/step
func (h *WorkflowHandler) AddWorkflowStep(c *gin.Context) {
	var req dto.AddWorkflowStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	_, err := h.service.AddWorkflowStep(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "status ticket is already in this workflow" {
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to add workflow step", err.Error())
		return
	}
	util.SuccessResponse(c, http.StatusCreated, gin.H{"message": "Workflow step added successfully"})
}

// GET ALL
func (h *WorkflowHandler) GetAllWorkflows(c *gin.Context) {
	workflows, err := h.service.GetAllWorkflows()
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve workflows", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, workflows)
}

func (h *WorkflowHandler) GetAllWorkflowSteps(c *gin.Context) {
	workflowIDStr := c.Query("workflow_id")
	if workflowIDStr != "" {
		workflowID, err := strconv.Atoi(workflowIDStr)
		if err != nil {
			util.ErrorResponse(c, http.StatusBadRequest, "Invalid workflow_id format", nil)
			return
		}
		steps, err := h.service.GetWorkflowStepsByWorkflowID(workflowID)
		if err != nil {
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve workflow steps", nil)
			return
		}
		util.SuccessResponse(c, http.StatusOK, steps)
		return
	}

	steps, err := h.service.GetAllWorkflowSteps()
	if err != nil {
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve all workflow steps", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, steps)
}

// GET BY ID
func (h *WorkflowHandler) GetWorkflowByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	workflow, err := h.service.GetWorkflowByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Workflow not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve workflow", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, workflow)
}

func (h *WorkflowHandler) GetWorkflowStepByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	step, err := h.service.GetWorkflowStepByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Workflow step not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve workflow step", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, step)
}

// UPDATE WORKFLOW NAME
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req dto.UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	updatedWorkflow, err := h.service.UpdateWorkflowName(id, req)
	if err != nil {
		switch err.Error() {
		case "workflow name already exists":
			util.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		case sql.ErrNoRows.Error():
			util.ErrorResponse(c, http.StatusNotFound, "Workflow not found", nil)
			return
		default:
			util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update workflow", nil)
			return
		}
	}
	util.SuccessResponse(c, http.StatusOK, updatedWorkflow)
}

// DELETE WORKFLOW
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.DeleteWorkflow(id); err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Workflow not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete workflow", nil)
		return
	}
	c.Status(http.StatusNoContent)
}

// DELETE WORKFLOW STEP
func (h *WorkflowHandler) DeleteWorkflowStep(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.DeleteWorkflowStep(id); err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Workflow step not found or sequence is not 0", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete workflow step", nil)
		return
	}
	c.Status(http.StatusNoContent)
}

// CHANGE WORKFLOW STATUS
func (h *WorkflowHandler) UpdateWorkflowActiveStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req dto.UpdateWorkflowStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := h.service.UpdateWorkflowActiveStatus(c.Request.Context(), id, req); err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Workflow not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update workflow status", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Workflow and its steps status updated successfully"})
}

// CHANGE WORKFLOW STEP STATUS
func (h *WorkflowHandler) UpdateWorkflowStepActiveStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req dto.UpdateWorkflowStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := h.service.UpdateWorkflowStepActiveStatus(id, req); err != nil {
		if err == sql.ErrNoRows {
			util.ErrorResponse(c, http.StatusNotFound, "Workflow step not found", nil)
			return
		}
		util.ErrorResponse(c, http.StatusInternalServerError, "Failed to update workflow step status", nil)
		return
	}
	util.SuccessResponse(c, http.StatusOK, gin.H{"message": "Workflow step status updated successfully"})
}

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"
	"e-memo-job-reservation-api/internal/websocket"

	"github.com/lib/pq"
)

type TicketWorkflowService struct {
	db                    *sql.DB
	ticketRepo            *repository.TicketRepository
	jobRepo               *repository.JobRepository
	employeeRepo          *repository.EmployeeRepository
	trackStatusTicketRepo *repository.TrackStatusTicketRepository
	statusTicketRepo      *repository.StatusTicketRepository
	statusTransitionRepo  *repository.StatusTransitionRepository
	actorRoleRepo         *repository.ActorRoleRepository
	actorRoleMappingRepo  *repository.ActorRoleMappingRepository
	ticketActionLogRepo   *repository.TicketActionLogRepository
	workflowRepo          *repository.WorkflowRepository
	actionService         *TicketActionService
	hub                   *websocket.Hub
	queryService          *TicketQueryService
}

type TicketWorkflowServiceConfig struct {
	DB                    *sql.DB
	TicketRepo            *repository.TicketRepository
	JobRepo               *repository.JobRepository
	EmployeeRepo          *repository.EmployeeRepository
	TrackStatusTicketRepo *repository.TrackStatusTicketRepository
	StatusTicketRepo      *repository.StatusTicketRepository
	StatusTransitionRepo  *repository.StatusTransitionRepository
	ActorRoleRepo         *repository.ActorRoleRepository
	ActorRoleMappingRepo  *repository.ActorRoleMappingRepository
	TicketActionLogRepo   *repository.TicketActionLogRepository
	WorkflowRepo          *repository.WorkflowRepository
	ActionService         *TicketActionService
	Hub                   *websocket.Hub
	QueryService          *TicketQueryService
}

func NewTicketWorkflowService(cfg *TicketWorkflowServiceConfig) *TicketWorkflowService {
	return &TicketWorkflowService{
		db:                    cfg.DB,
		ticketRepo:            cfg.TicketRepo,
		jobRepo:               cfg.JobRepo,
		employeeRepo:          cfg.EmployeeRepo,
		trackStatusTicketRepo: cfg.TrackStatusTicketRepo,
		statusTicketRepo:      cfg.StatusTicketRepo,
		statusTransitionRepo:  cfg.StatusTransitionRepo,
		actorRoleRepo:         cfg.ActorRoleRepo,
		actorRoleMappingRepo:  cfg.ActorRoleMappingRepo,
		ticketActionLogRepo:   cfg.TicketActionLogRepo,
		workflowRepo:          cfg.WorkflowRepo,
		actionService:         cfg.ActionService,
		hub:                   cfg.Hub,
		queryService:          cfg.QueryService,
	}
}

// EXECUTE ACTION TO GET TO THE NEXT STATUS BASED ON STATE
func (s *TicketWorkflowService) ExecuteAction(ctx context.Context, ticketID int, userNPK string, req dto.ExecuteActionRequest, filesMetadata []model.FileMetadata) error {
	availableActions, err := s.actionService.GetAvailableActions(ctx, ticketID, userNPK)
	if err != nil {
		return err
	}

	var selectedAction *dto.AvailableTicketActionResponse
	for _, action := range availableActions {
		if action.ActionName == req.ActionName {
			act := action
			selectedAction = &act
			break
		}
	}

	if selectedAction == nil {
		return errors.New("user does not have the required role or action is not allowed from the current status")
	}

	if selectedAction.RequireReason && req.Reason == "" {
		errorMsg := "reason is required for this action"
		if selectedAction.ReasonLabel != nil {
			errorMsg = fmt.Sprintf("%s is required", *selectedAction.ReasonLabel)
		}
		return errors.New(errorMsg)
	}

	if selectedAction.RequireFile && len(filesMetadata) == 0 {
		return errors.New("file upload is required for this action")
	}

	var finalToStatusID int

	if req.ActionName == "Revisi" {
		user, err := s.employeeRepo.FindByNPK(userNPK)
		if err != nil {
			return errors.New("user not found")
		}

		initialStatusID, err := s.workflowRepo.GetInitialStatusByPosition(ctx, user.Position.ID)
		if err != nil {
			return errors.New("no workflow defined for this user's position")
		}
		finalToStatusID = initialStatusID
	} else {
		finalToStatusID = selectedAction.ToStatusID
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var oldReportFiles []model.FileMetadata

	if req.ActionName == "Selesaikan Job" {
		if len(filesMetadata) > 0 {
			// Get existing report files before updating (for cleanup after commit)
			existingJob, err := s.jobRepo.FindByTicketID(ctx, ticketID)
			if err != nil && err != sql.ErrNoRows {
				return errors.New("failed to retrieve existing job data")
			}
			if existingJob != nil {
				oldReportFiles = existingJob.ReportFiles
			}

			err = s.jobRepo.UpdateJobCompletionDetails(ctx, tx, ticketID, filesMetadata, req.SpendingAmount)
			if err != nil {
				if err == sql.ErrNoRows {
					return errors.New("job associated with this ticket not found")
				}
				return errors.New("failed to update job completion details")
			}
		}
	}

	currentStatusID, _, _ := s.trackStatusTicketRepo.GetCurrentStatusByTicketID(ctx, ticketID)

	var filePathsForLog []string
	for _, meta := range filesMetadata {
		filePathsForLog = append(filePathsForLog, meta.FilePath)
	}

	logEntry := model.TicketActionLog{
		TicketID:       int64(ticketID),
		ActionID:       selectedAction.ActionID,
		PerformedByNpk: userNPK,
		DetailsText:    sql.NullString{String: req.Reason, Valid: req.Reason != ""},
		FilePath:       pq.StringArray(filePathsForLog),
		FromStatusID:   sql.NullInt32{Int32: int32(currentStatusID), Valid: true},
		ToStatusID:     finalToStatusID,
	}
	if err := s.ticketActionLogRepo.Create(ctx, tx, logEntry); err != nil {
		return err
	}

	if err := s.trackStatusTicketRepo.UpdateStatus(ctx, tx, ticketID, finalToStatusID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// After successful commit, clean up old report files from storage
	if req.ActionName == "Selesaikan Job" && len(filesMetadata) > 0 && len(oldReportFiles) > 0 {
		for _, oldFile := range oldReportFiles {
			if err := os.Remove(oldFile.FilePath); err != nil {
				log.Printf("WARNING: Failed to delete old report file from storage. File path: %s, Error: %v", oldFile.FilePath, err)
			}
		}
	}

	updatedTicketDetail, err := s.queryService.GetTicketByID(ticketID)
	if err != nil {
		log.Printf("CRITICAL: Failed to fetch updated ticket for broadcast after action. TicketID: %d, Error: %v", ticketID, err)
	} else {
		message, err := websocket.NewMessage("TICKET_STATUS_CHANGED", updatedTicketDetail)
		if err != nil {
			log.Printf("CRITICAL: Failed to create websocket message for status change: %v", err)
		} else {
			s.hub.BroadcastMessage(message)
		}
	}

	return nil
}

func (s *TicketWorkflowService) ValidateAndGetTransition(ctx context.Context, currentStatusID int, actionName string) (toStatusID int, allowedRoleIDs []int, err error) {
	toStatusID, allowedRoleIDs, err = s.statusTransitionRepo.FindValidTransition(currentStatusID, actionName)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil, errors.New("action not allowed from the current status")
		}
		return 0, nil, err
	}

	if len(allowedRoleIDs) == 0 {
		return 0, nil, errors.New("no roles are configured to perform this action")
	}

	return toStatusID, allowedRoleIDs, nil
}

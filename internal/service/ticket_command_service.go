package service

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"
	"e-memo-job-reservation-api/internal/websocket"
	"e-memo-job-reservation-api/pkg/filehandler"

	"github.com/gin-gonic/gin"
)

type TicketCommandService struct {
	db                    *sql.DB
	ticketRepo            *repository.TicketRepository
	jobRepo               *repository.JobRepository
	workflowRepo          *repository.WorkflowRepository
	trackStatusTicketRepo *repository.TrackStatusTicketRepository
	employeeRepo          *repository.EmployeeRepository
	departmentRepo        *repository.DepartmentRepository
	actorRoleMappingRepo  *repository.ActorRoleMappingRepository
	actorRoleRepo         *repository.ActorRoleRepository
	statusTransitionRepo  *repository.StatusTransitionRepository
	specifiedLocationRepo *repository.SpecifiedLocationRepository
	hub                   *websocket.Hub
	queryService          *TicketQueryService
}

type TicketCommandServiceConfig struct {
	DB                    *sql.DB
	TicketRepo            *repository.TicketRepository
	JobRepo               *repository.JobRepository
	WorkflowRepo          *repository.WorkflowRepository
	TrackStatusTicketRepo *repository.TrackStatusTicketRepository
	EmployeeRepo          *repository.EmployeeRepository
	DepartmentRepo        *repository.DepartmentRepository
	ActorRoleMappingRepo  *repository.ActorRoleMappingRepository
	ActorRoleRepo         *repository.ActorRoleRepository
	StatusTransitionRepo  *repository.StatusTransitionRepository
	SpecifiedLocationRepo *repository.SpecifiedLocationRepository
	Hub                   *websocket.Hub
	QueryService          *TicketQueryService
}

func NewTicketCommandService(cfg *TicketCommandServiceConfig) *TicketCommandService {
	return &TicketCommandService{
		db:                    cfg.DB,
		ticketRepo:            cfg.TicketRepo,
		jobRepo:               cfg.JobRepo,
		workflowRepo:          cfg.WorkflowRepo,
		trackStatusTicketRepo: cfg.TrackStatusTicketRepo,
		employeeRepo:          cfg.EmployeeRepo,
		departmentRepo:        cfg.DepartmentRepo,
		actorRoleMappingRepo:  cfg.ActorRoleMappingRepo,
		actorRoleRepo:         cfg.ActorRoleRepo,
		statusTransitionRepo:  cfg.StatusTransitionRepo,
		specifiedLocationRepo: cfg.SpecifiedLocationRepo,
		hub:                   cfg.Hub,
		queryService:          cfg.QueryService,
	}
}

// CREATE TICKET
func (s *TicketCommandService) CreateTicket(ctx context.Context, req dto.CreateTicketRequest, requestor string, filesMetadata []model.FileMetadata) (*model.Ticket, error) { // VALIDATE DEPARTMENT
	canReceive, err := s.departmentRepo.IsReceiver(req.DepartmentTargetID)
	if err != nil {
		return nil, err // "department not found or is not active"
	}
	if !canReceive {
		return nil, errors.New("selected target department cannot receive jobs")
	}

	// GET EMPLOYEE DATA (TO GET THE POSITION)
	positionID, err := s.employeeRepo.GetEmployeePositionID(ctx, requestor)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("requestor not found")
		}
		return nil, err
	}

	// GET INITIAL STATUS FROM PREVIOUS GET EMPLOYEE DATA
	initialStatusID, err := s.workflowRepo.GetInitialStatusByPosition(ctx, positionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no workflow defined for this user's position")
		}
		return nil, err
	}

	deadline, err := repository.ParseDeadline(req.Deadline)
	if err != nil {
		return nil, errors.New("invalid deadline format, please use YYYY-MM-DD")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var specifiedLocationID sql.NullInt64
	if req.SpecifiedLocationName != nil && req.PhysicalLocationID != nil {
		id, err := s.specifiedLocationRepo.FindOrCreate(ctx, tx, *req.SpecifiedLocationName, *req.PhysicalLocationID)
		if err != nil {
			return nil, errors.New("failed to find or create specified location")
		}
		specifiedLocationID = toNullInt64(&id)
	}

	// GET LAST PRIORITY
	lastPriority, err := s.ticketRepo.GetLastPriority(ctx, tx, req.DepartmentTargetID)
	if err != nil {
		return nil, err
	}

	ticketData := model.Ticket{
		Requestor:           requestor,
		DepartmentTargetID:  req.DepartmentTargetID,
		PhysicalLocationID:  toNullInt64(req.PhysicalLocationID),
		SpecifiedLocationID: specifiedLocationID,
		Description:         req.Description,
		TicketPriority:      lastPriority,
		Deadline:            deadline,
		SupportFiles:        filesMetadata,
	}

	// INSERT DATA TO TICKET TABLE
	createdTicket, err := s.ticketRepo.Create(ctx, tx, ticketData)
	if err != nil {
		return nil, err
	}

	// INSERT DATA TO JOB TABLE
	err = s.jobRepo.Create(ctx, tx, createdTicket.ID, createdTicket.TicketPriority)
	if err != nil {
		return nil, err
	}

	// INITIATE FIRST STATUS
	err = s.trackStatusTicketRepo.CreateInitialStatus(ctx, tx, createdTicket.ID, initialStatusID)
	if err != nil {
		return nil, err
	}

	// COMMIT DATA
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	ticketDetail, err := s.queryService.GetTicketByID(createdTicket.ID)
	if err != nil {
		log.Printf("CRITICAL: Failed to fetch new ticket details for broadcast. TicketID: %d, Error: %v", createdTicket.ID, err)
	} else {
		message, err := websocket.NewMessage("TICKET_CREATED", ticketDetail)
		if err != nil {
			log.Printf("CRITICAL: Failed to create websocket message for new ticket: %v", err)
		} else {
			s.hub.BroadcastMessage(message)
		}
	}

	return createdTicket, err
}

// UPDATE TICKET
func (s *TicketCommandService) UpdateTicket(ctx context.Context, ticketID int, req dto.UpdateTicketRequest, userNPK string) error {
	originalTicket, err := s.ticketRepo.FindByIDAsStruct(ctx, ticketID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("ticket not found")
		}
		return err
	}

	user, err := s.employeeRepo.FindByNPK(userNPK)
	if err != nil {
		return errors.New("user not found")
	}

	requestor, err := s.employeeRepo.FindByNPK(originalTicket.Requestor)
	if err != nil {
		return errors.New("original requestor not found")
	}

	currentStatusID, _, err := s.trackStatusTicketRepo.GetCurrentStatusByTicketID(ctx, ticketID)
	if err != nil {
		return errors.New("could not retrieve current ticket status")
	}

	userContexts := determineUserContexts(user, originalTicket, requestor, nil)

	actorRoles, err := s.actorRoleMappingRepo.GetRolesForUserContext(user.Position.ID, userContexts)
	if err != nil {
		return err
	}

	userActorRoleIDs, err := s.actorRoleRepo.GetRoleIDsByNames(actorRoles)
	if err != nil {
		return err
	}

	_, allowedRoleIDsForRevise, err := s.statusTransitionRepo.FindValidTransition(currentStatusID, "Revisi")

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("ticket cannot be edited in its current state")
		}
		return err
	}

	isAuthorized := false
	for _, userRoleID := range userActorRoleIDs {
		for _, allowedRoleID := range allowedRoleIDsForRevise {
			if userRoleID == allowedRoleID {
				isAuthorized = true
				break
			}
		}
		if isAuthorized {
			break
		}
	}

	if !isAuthorized {
		return errors.New("user is not authorized to edit this ticket")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var specifiedLocationID sql.NullInt64
	if req.SpecifiedLocationName != nil && req.PhysicalLocationID != nil {
		id, err := s.specifiedLocationRepo.FindOrCreate(ctx, tx, *req.SpecifiedLocationName, *req.PhysicalLocationID)
		if err != nil {
			return errors.New("failed to find or create specified location")
		}
		specifiedLocationID = toNullInt64(&id)
	}

	rowsAffected, err := s.ticketRepo.Update(ctx, tx, ticketID, req, specifiedLocationID)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.New("invalid deadline format, please use YYYY-MM-DD")
		}
		return err
	}
	if rowsAffected == 0 {
		return errors.New("data conflict: ticket has been modified by another user, please refresh")
	}

	if req.Deadline != nil {
		if _, err := time.Parse("2006-01-02", *req.Deadline); err != nil {
			return errors.New("invalid deadline format, please use YYYY-MM-DD")
		}
	}

	updatedTicketDetail, err := s.queryService.GetTicketByID(ticketID)
	if err != nil {
		log.Printf("CRITICAL: Failed to fetch updated ticket for broadcast after update. TicketID: %d, Error: %v", ticketID, err)
	} else {
		message, err := websocket.NewMessage("TICKET_UPDATED", updatedTicketDetail)
		if err != nil {
			log.Printf("CRITICAL: Failed to create websocket message for updated ticket: %v", err)
		} else {
			s.hub.BroadcastMessage(message)
		}
	}

	return tx.Commit()
}

func (s *TicketCommandService) AddSupportFiles(ctx context.Context, c *gin.Context, ticketID int, userNPK string, files []*multipart.FileHeader) error {
	originalTicket, err := s.ticketRepo.FindByIDAsStruct(ctx, ticketID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("ticket not found")
		}
		return err
	}

	user, err := s.employeeRepo.FindByNPK(userNPK)
	if err != nil {
		return errors.New("user not found")
	}

	requestor, err := s.employeeRepo.FindByNPK(originalTicket.Requestor)
	if err != nil {
		return errors.New("original requestor not found")
	}

	currentStatusID, _, err := s.trackStatusTicketRepo.GetCurrentStatusByTicketID(ctx, ticketID)
	if err != nil {
		return errors.New("could not retrieve current ticket status")
	}

	userContexts := determineUserContexts(user, originalTicket, requestor, nil)

	actorRoles, err := s.actorRoleMappingRepo.GetRolesForUserContext(user.Position.ID, userContexts)
	if err != nil {
		return err
	}

	userActorRoleIDs, err := s.actorRoleRepo.GetRoleIDsByNames(actorRoles)
	if err != nil {
		return err
	}

	_, allowedRoleIDsForRevise, err := s.statusTransitionRepo.FindValidTransition(currentStatusID, "Revisi")

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("ticket cannot be edited in its current state")
		}
		return err
	}

	isAuthorized := false
	for _, userRoleID := range userActorRoleIDs {
		for _, allowedRoleID := range allowedRoleIDsForRevise {
			if userRoleID == allowedRoleID {
				isAuthorized = true
				break
			}
		}
		if isAuthorized {
			break
		}
	}

	if !isAuthorized {
		return errors.New("user is not authorized to edit this ticket")
	}

	savedFilesMetadata, err := filehandler.SaveFiles(c, files)
	if err != nil {
		return errors.New("failed to save one or more files")
	}

	if len(savedFilesMetadata) == 0 {
		return nil
	}

	if err := s.ticketRepo.AddSupportFiles(ctx, ticketID, savedFilesMetadata); err != nil {
		for _, metadata := range savedFilesMetadata {
			os.Remove(metadata.FilePath)
		}
		return err
	}

	updatedTicketDetail, err := s.queryService.GetTicketByID(ticketID)
	if err != nil {
		log.Printf("CRITICAL: Failed to fetch updated ticket for broadcast after adding files. TicketID: %d, Error: %v", ticketID, err)
	} else {
		message, err := websocket.NewMessage("TICKET_UPDATED", updatedTicketDetail)
		if err != nil {
			log.Printf("CRITICAL: Failed to create websocket message for updated ticket files: %v", err)
		} else {
			s.hub.BroadcastMessage(message)
		}
	}

	return nil
}

func (s *TicketCommandService) RemoveSupportFiles(ctx context.Context, ticketID int, userNPK string, req dto.DeleteFilesRequest) error {
	originalTicket, err := s.ticketRepo.FindByIDAsStruct(ctx, ticketID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("ticket not found")
		}
		return err
	}

	user, err := s.employeeRepo.FindByNPK(userNPK)
	if err != nil {
		return errors.New("user not found")
	}

	requestor, err := s.employeeRepo.FindByNPK(originalTicket.Requestor)
	if err != nil {
		return errors.New("original requestor not found")
	}

	currentStatusID, _, err := s.trackStatusTicketRepo.GetCurrentStatusByTicketID(ctx, ticketID)
	if err != nil {
		return errors.New("could not retrieve current ticket status")
	}

	userContexts := determineUserContexts(user, originalTicket, requestor, nil)

	actorRoles, err := s.actorRoleMappingRepo.GetRolesForUserContext(user.Position.ID, userContexts)
	if err != nil {
		return err
	}

	userActorRoleIDs, err := s.actorRoleRepo.GetRoleIDsByNames(actorRoles)
	if err != nil {
		return err
	}

	_, allowedRoleIDsForRevise, err := s.statusTransitionRepo.FindValidTransition(currentStatusID, "Revisi")

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("ticket cannot be edited in its current state")
		}
		return err
	}

	isAuthorized := false
	for _, userRoleID := range userActorRoleIDs {
		for _, allowedRoleID := range allowedRoleIDsForRevise {
			if userRoleID == allowedRoleID {
				isAuthorized = true
				break
			}
		}
		if isAuthorized {
			break
		}
	}

	if !isAuthorized {
		return errors.New("user is not authorized to edit this ticket")
	}

	if err := s.ticketRepo.RemoveSupportFiles(ctx, ticketID, req.FilePathsToDelete); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("ticket not found")
		}
		return err
	}

	for _, filePath := range req.FilePathsToDelete {
		if err := os.Remove(filePath); err != nil {
			log.Printf("WARNING: Failed to delete file from storage, but DB record was removed. File path: %s, Error: %v", filePath, err)
		}
	}

	updatedTicketDetail, err := s.queryService.GetTicketByID(ticketID)
	if err != nil {
		log.Printf("CRITICAL: Failed to fetch updated ticket for broadcast after removing files. TicketID: %d, Error: %v", ticketID, err)
	} else {
		message, err := websocket.NewMessage("TICKET_UPDATED", updatedTicketDetail)
		if err != nil {
			log.Printf("CRITICAL: Failed to create websocket message for removed ticket files: %v", err)
		} else {
			s.hub.BroadcastMessage(message)
		}
	}

	return nil
}

// HELPER
// CONVERTER
func toNullInt64(val *int) sql.NullInt64 {
	if val == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: int64(*val), Valid: true}
}

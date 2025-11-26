package service

import (
	"context"
	"database/sql"
	"errors"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/repository"
)

type TicketActionService struct {
	ticketRepo            *repository.TicketRepository
	jobRepo               *repository.JobRepository
	employeeRepo          *repository.EmployeeRepository
	trackStatusTicketRepo *repository.TrackStatusTicketRepository
	statusTransitionRepo  *repository.StatusTransitionRepository
	actorRoleMappingRepo  *repository.ActorRoleMappingRepository
	actorRoleRepo         *repository.ActorRoleRepository
}

type TicketActionServiceConfig struct {
	TicketRepo            *repository.TicketRepository
	JobRepo               *repository.JobRepository
	EmployeeRepo          *repository.EmployeeRepository
	TrackStatusTicketRepo *repository.TrackStatusTicketRepository
	StatusTransitionRepo  *repository.StatusTransitionRepository
	ActorRoleMappingRepo  *repository.ActorRoleMappingRepository
	ActorRoleRepo         *repository.ActorRoleRepository
}

func NewTicketActionService(cfg *TicketActionServiceConfig) *TicketActionService {
	return &TicketActionService{
		ticketRepo:            cfg.TicketRepo,
		jobRepo:               cfg.JobRepo,
		employeeRepo:          cfg.EmployeeRepo,
		trackStatusTicketRepo: cfg.TrackStatusTicketRepo,
		statusTransitionRepo:  cfg.StatusTransitionRepo,
		actorRoleMappingRepo:  cfg.ActorRoleMappingRepo,
		actorRoleRepo:         cfg.ActorRoleRepo,
	}
}

// GET AVAILABLE ACTIONS
func (s *TicketActionService) GetAvailableActions(ctx context.Context, ticketID int, userNPK string) ([]dto.AvailableTicketActionResponse, error) {
	user, err := s.employeeRepo.FindByNPK(userNPK)
	if err != nil {
		return nil, errors.New("user employee not found")
	}

	ticket, err := s.ticketRepo.FindByIDAsStruct(ctx, ticketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	requestor, err := s.employeeRepo.FindByNPK(ticket.Requestor)
	if err != nil {
		return nil, errors.New("requestor employee not found")
	}

	job, err := s.jobRepo.FindByTicketID(ctx, ticketID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	userContexts := determineUserContexts(user, ticket, requestor, job)

	userRoleIDs, err := s.actorRoleMappingRepo.GetRoleIDsForUserContext(user.Position.ID, userContexts)
	if err != nil {
		return nil, err
	}

	if job != nil && job.PicJob.Valid && user.NPK == job.PicJob.String {
		assignedPicRoleIDs, err := s.actorRoleRepo.GetRoleIDsByNames([]string{"ASSIGNED_PIC"})
		if err != nil {
			return nil, err
		} else {
			userRoleIDs = append(userRoleIDs, assignedPicRoleIDs...)
		}
	}

	currentStatusID, _, err := s.trackStatusTicketRepo.GetCurrentStatusByTicketID(ctx, ticketID)
	if err != nil {
		return nil, errors.New("current status not found")
	}

	availableActions, err := s.statusTransitionRepo.FindAvailableTransitionsForRoles(currentStatusID, userRoleIDs)
	if err != nil {
		return nil, err
	}

	return availableActions, nil
}

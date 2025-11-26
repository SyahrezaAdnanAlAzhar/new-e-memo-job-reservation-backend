package service

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/repository"
	"e-memo-job-reservation-api/internal/websocket"

	"github.com/gin-gonic/gin"
)

type JobService struct {
	jobCommandRepo *repository.JobRepository
	jobQueryRepo   *repository.JobQueryRepository
	employeeRepo   *repository.EmployeeRepository
	posPermRepo    *repository.PositionPermissionRepository
	db             *sql.DB
	hub            *websocket.Hub
	queryService   *TicketQueryService
}

func NewJobService(jobCommandRepo *repository.JobRepository, jobQueryRepo *repository.JobQueryRepository, employeeRepo *repository.EmployeeRepository, posPermRepo *repository.PositionPermissionRepository, db *sql.DB, hub *websocket.Hub, queryService *TicketQueryService) *JobService {
	return &JobService{
		jobCommandRepo: jobCommandRepo,
		jobQueryRepo:   jobQueryRepo,
		employeeRepo:   employeeRepo,
		posPermRepo:    posPermRepo,
		db:             db,
		hub:            hub,
		queryService:   queryService,
	}
}

// AssignPIC
func (s *JobService) AssignPIC(ctx context.Context, jobID int, req dto.AssignPICRequest, userNPK string) error {
	user, err := s.employeeRepo.FindByNPK(userNPK)
	if err != nil {
		return errors.New("action performer not found")
	}

	job, err := s.jobQueryRepo.FindByID(jobID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("job not found")
		}
		return err
	}

	newPIC, err := s.employeeRepo.FindByNPK(req.PicJob)
	if err != nil {
		return errors.New("new PIC employee data not found")
	}

	if user.DepartmentID != job.AssignedDepartmentID {
		return errors.New("user is not authorized to assign PIC for this job's department")
	}

	if newPIC.DepartmentID != job.AssignedDepartmentID {
		return errors.New("new PIC must be from the same department as the job")
	}

	err = s.jobCommandRepo.AssignPIC(jobID, req.PicJob)
	if err != nil {
		return err
	}

	job, _ = s.jobQueryRepo.FindByID(jobID)
	if job != nil {
		updatedTicketDetail, err := s.queryService.GetTicketByID(job.TicketID)
		if err != nil {
			log.Printf("CRITICAL: Failed to fetch updated ticket for broadcast after PIC assign. TicketID: %d, Error: %v", job.TicketID, err)
		} else {
			message, err := websocket.NewMessage("TICKET_UPDATED", updatedTicketDetail)
			if err != nil {
				log.Printf("CRITICAL: Failed to create websocket message for PIC assign: %v", err)
			} else {
				s.hub.BroadcastMessage(message)
			}
		}
	}

	return nil
}

// ReorderJobs
func (s *JobService) ReorderJobs(ctx context.Context, req dto.ReorderJobsRequest, userNPK string) error {
	user, err := s.employeeRepo.FindByNPK(userNPK)
	if err != nil {
		return errors.New("action performer not found")
	}
	if user.DepartmentID != req.DepartmentTargetID {
		return errors.New("user can only reorder jobs within their own department")
	}

	jobIDs := make([]int, len(req.Items))
	for i, item := range req.Items {
		jobIDs[i] = item.JobID
	}

	validJobCount, err := s.jobCommandRepo.CheckJobsInDepartment(jobIDs, req.DepartmentTargetID)
	if err != nil {
		return err
	}
	if validJobCount != len(req.Items) {
		return errors.New("one or more job IDs do not belong to the specified department")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, item := range req.Items {
		newPriority := i + 1
		rowsAffected, err := s.jobCommandRepo.UpdatePriority(ctx, tx, item.JobID, item.Version, newPriority)
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return errors.New("data conflict: job has been modified by another user, please refresh")
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	payload := gin.H{
		"department_target_id": req.DepartmentTargetID,
		"message":              "Job priorities have been updated.",
	}
	message, err := websocket.NewMessage("JOB_PRIORITY_UPDATED", payload)
	if err != nil {
		log.Printf("CRITICAL: Failed to create websocket message for job reorder: %v", err)
	} else {
		s.hub.BroadcastMessage(message)
	}

	return nil
}

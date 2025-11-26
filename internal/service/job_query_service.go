package service

import (
	"context"
	"errors"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/repository"
)

type JobQueryService struct {
	queryRepo    *repository.JobQueryRepository
	employeeRepo *repository.EmployeeRepository
	posPermRepo  *repository.PositionPermissionRepository
}

func NewJobQueryService(queryRepo *repository.JobQueryRepository, employeeRepo *repository.EmployeeRepository, posPermRepo *repository.PositionPermissionRepository) *JobQueryService {
	return &JobQueryService{
		queryRepo:    queryRepo,
		employeeRepo: employeeRepo,
		posPermRepo:  posPermRepo,
	}
}

// GET ALL
func (s *JobQueryService) GetAllJobs(filters dto.JobFilter) ([]dto.JobDetailResponse, error) {
	return s.queryRepo.FindAll(filters)
}

// GET BY ID
func (s *JobQueryService) GetJobByID(id int) (*dto.JobDetailResponse, error) {
	return s.queryRepo.FindByID(id)
}

func (s *JobQueryService) GetAvailableActions(ctx context.Context, jobID int, userNPK string) ([]dto.AvailableActionResponse, error) {
	user, err := s.employeeRepo.FindByNPK(userNPK)
	if err != nil {
		return nil, errors.New("user not found")
	}

	job, err := s.queryRepo.FindByID(jobID)
	if err != nil {
		return nil, errors.New("job not found")
	}

	allPermissions, err := s.posPermRepo.FindPermissionsByPositionID(user.Position.ID)
	if err != nil {
		return nil, err
	}

	var availableActions []dto.AvailableActionResponse
	for _, p := range allPermissions {
		isActionAllowed := false
		switch p.Name {
		case "JOB_ASSIGN_PIC":
			if user.DepartmentID == job.AssignedDepartmentID {
				isActionAllowed = true
			}
		}

		if isActionAllowed {
			availableActions = append(availableActions, p)
		}
	}

	return availableActions, nil
}

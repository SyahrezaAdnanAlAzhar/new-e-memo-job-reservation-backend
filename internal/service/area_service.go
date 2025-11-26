package service

import (
	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type AreaService struct {
	repo *repository.AreaRepository
}

func NewAreaService(repo *repository.AreaRepository) *AreaService {
	return &AreaService{repo: repo}
}

// CREATE
func (s *AreaService) CreateArea(req dto.CreateAreaRequest) (*model.Area, error) {
	if req.Name == "" {
		return nil, errors.New("area name is required")
	}
	if req.DepartmentID <= 0 {
		return nil, errors.New("valid department_id is required")
	}

	newArea, err := s.repo.Create(req)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New("area name already exists in this department")
		}
		return nil, err
	}

	return newArea, nil
}

// GET ALL
func (s *AreaService) GetAllAreas(filters map[string]string) ([]model.Area, error) {
	return s.repo.FindAll(filters)
}

// GET BY ID
func (s *AreaService) GetAreaByID(id int) (*model.Area, error) {
	return s.repo.FindByID(id)
}

// DELETE
func (s *AreaService) DeleteArea(id int) error {
	return s.repo.Delete(id)
}

// UPDATE
func (s *AreaService) UpdateArea(id int, req dto.UpdateAreaRequest) (*model.Area, error) {
	isTaken, err := s.repo.IsNameTakenInDepartment(req.Name, req.DepartmentID, id)
	if err != nil {
		return nil, err
	}
	if isTaken {
		return nil, errors.New("area name already exists in this department")
	}

	return s.repo.Update(id, req)
}

// CHANGE ACTIVE STATUS
func (s *AreaService) UpdateAreaActiveStatus(id int, req repository.UpdateAreaStatusRequest) error {
	return s.repo.UpdateActiveStatus(id, req.IsActive)
}

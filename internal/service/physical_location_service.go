package service

import (
	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type PhysicalLocationService struct {
	repo *repository.PhysicalLocationRepository
}

func NewPhysicalLocationService(repo *repository.PhysicalLocationRepository) *PhysicalLocationService {
	return &PhysicalLocationService{repo: repo}
}

// CREATE
func (s *PhysicalLocationService) CreatePhysicalLocation(req dto.CreatePhysicalLocationRequest) (*model.PhysicalLocation, error) {
	newLoc, err := s.repo.Create(req)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New("physical location name already exists")
		}
		return nil, err
	}
	return newLoc, nil
}

// GET ALL
func (s *PhysicalLocationService) GetAllPhysicalLocations(filters map[string]string) ([]model.PhysicalLocation, error) {
	return s.repo.FindAll(filters)
}

// GET BY ID
func (s *PhysicalLocationService) GetPhysicalLocationByID(id int) (*model.PhysicalLocation, error) {
	return s.repo.FindByID(id)
}

// UPDATE
func (s *PhysicalLocationService) UpdatePhysicalLocation(id int, req dto.UpdatePhysicalLocationRequest) (*model.PhysicalLocation, error) {
	isTaken, err := s.repo.IsNameTaken(req.Name, id)
	if err != nil {
		return nil, err // Teruskan error database
	}
	if isTaken {
		return nil, errors.New("physical location name already exists")
	}

	return s.repo.Update(id, req)
}

// DELETE
func (s *PhysicalLocationService) DeletePhysicalLocation(id int) error {
	return s.repo.Delete(id)
}

// CHANGE STATUS
func (s *PhysicalLocationService) UpdatePhysicalLocationActiveStatus(id int, req dto.UpdatePhysicalLocationStatusRequest) error {
	return s.repo.UpdateActiveStatus(id, req.IsActive)
}

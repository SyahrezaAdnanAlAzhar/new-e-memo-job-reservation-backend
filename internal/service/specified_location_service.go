package service

import (
	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type SpecifiedLocationService struct {
	repo *repository.SpecifiedLocationRepository
}

func NewSpecifiedLocationService(repo *repository.SpecifiedLocationRepository) *SpecifiedLocationService {
	return &SpecifiedLocationService{repo: repo}
}

// CREATE
func (s *SpecifiedLocationService) CreateSpecifiedLocation(req dto.CreateSpecifiedLocationRequest) (*model.SpecifiedLocation, error) {
	newLoc, err := s.repo.Create(req)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" { // foreign_key_violation
				return nil, errors.New("invalid physical_location_id")
			}
			if pgErr.Code == "23505" { // unique_violation
				return nil, errors.New("location name already exists in this physical location")
			}
		}
		return nil, err
	}
	return newLoc, nil
}

// GET ALL
func (s *SpecifiedLocationService) GetAllSpecifiedLocations(filters dto.SpecifiedLocationFilter) ([]model.SpecifiedLocation, error) {
	return s.repo.FindAll(filters)
}

// GET ALL BY PHYSICAL LOCATION ID
func (s *SpecifiedLocationService) GetSpecifiedLocationsByPhysicalLocationID(physicalLocationID int) ([]model.SpecifiedLocation, error) {
	return s.repo.FindByPhysicalLocationID(physicalLocationID)
}

// GET BY ID
func (s *SpecifiedLocationService) GetSpecifiedLocationByID(id int) (*model.SpecifiedLocation, error) {
	return s.repo.FindByID(id)
}

// UPDATE
func (s *SpecifiedLocationService) UpdateSpecifiedLocation(id int, req dto.UpdateSpecifiedLocationRequest) (*model.SpecifiedLocation, error) {
	isTaken, err := s.repo.IsNameTakenInPhysicalLocation(req.Name, req.PhysicalLocationID, id)
	if err != nil {
		return nil, err
	}
	if isTaken {
		return nil, errors.New("location name already exists in this physical location")
	}

	return s.repo.Update(id, req)
}

// DELETE
func (s *SpecifiedLocationService) DeleteSpecifiedLocation(id int) error {
	return s.repo.Delete(id)
}

// CHANGE STATUS
func (s *SpecifiedLocationService) UpdateSpecifiedLocationActiveStatus(id int, req dto.UpdateSpecifiedLocationStatusRequest) error {
	return s.repo.UpdateActiveStatus(id, req.IsActive)
}

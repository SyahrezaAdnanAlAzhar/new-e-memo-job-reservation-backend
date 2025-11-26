package service

import (
	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type AccessPermissionService struct {
	repo *repository.AccessPermissionRepository
}

func NewAccessPermissionService(repo *repository.AccessPermissionRepository) *AccessPermissionService {
	return &AccessPermissionService{repo: repo}
}

// CREATE
func (s *AccessPermissionService) CreateAccessPermission(req dto.CreateAccessPermissionRequest) (*model.AccessPermission, error) {
	newPermission, err := s.repo.Create(req)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New("permission name already exists")
		}
		return nil, err
	}
	return newPermission, nil
}

// GET ALL
func (s *AccessPermissionService) GetAllAccessPermissions() ([]model.AccessPermission, error) {
	return s.repo.FindAll()
}

// GET BY ID
func (s *AccessPermissionService) GetAccessPermissionByID(id int) (*model.AccessPermission, error) {
	return s.repo.FindByID(id)
}

// UPDATE
func (s *AccessPermissionService) UpdateAccessPermission(id int, req dto.UpdateAccessPermissionRequest) (*model.AccessPermission, error) {
	isTaken, err := s.repo.IsNameTaken(req.Name, id)
	if err != nil {
		return nil, err
	}
	if isTaken {
		return nil, errors.New("permission name already exists")
	}

	return s.repo.Update(id, req)
}

// DELETE
func (s *AccessPermissionService) DeleteAccessPermission(id int) error {
	return s.repo.Delete(id)
}

// CHANGE STATUS
func (s *AccessPermissionService) UpdateAccessPermissionActiveStatus(id int, req repository.UpdateAccessPermissionStatusRequest) error {
	return s.repo.UpdateActiveStatus(id, req.IsActive)
}

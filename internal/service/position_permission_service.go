package service

import (
	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type PositionPermissionService struct {
	repo *repository.PositionPermissionRepository
}

func NewPositionPermissionService(repo *repository.PositionPermissionRepository) *PositionPermissionService {
	return &PositionPermissionService{repo: repo}
}

// CREATE
func (s *PositionPermissionService) CreatePositionPermission(req dto.CreatePositionPermissionRequest) (*model.PositionPermission, error) {
	newPerm, err := s.repo.Create(req)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" { // foreign_key_violation
				return nil, errors.New("invalid employee_position_id or access_permission_id")
			}
			if pgErr.Code == "23505" { // unique_violation
				return nil, errors.New("this permission is already assigned to the position")
			}
		}
		return nil, err
	}
	return newPerm, nil
}

// GET ALL
func (s *PositionPermissionService) GetAllPositionPermissions() ([]model.PositionPermission, error) {
	return s.repo.FindAll()
}

// CHANGE STATUS
func (s *PositionPermissionService) UpdatePositionPermissionActiveStatus(posID, permID int, req dto.UpdatePositionPermissionStatusRequest) error {
	return s.repo.UpdateActiveStatus(posID, permID, req.IsActive)
}

// DELETE
func (s *PositionPermissionService) DeletePositionPermission(posID, permID int) error {
	return s.repo.Delete(posID, permID)
}

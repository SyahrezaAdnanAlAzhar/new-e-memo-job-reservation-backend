package service

import (
	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/repository"
)

type ActionService struct {
	repo *repository.ActionRepository
}

func NewActionService(repo *repository.ActionRepository) *ActionService {
	return &ActionService{repo: repo}
}

func (s *ActionService) GetAllActions() ([]dto.ActionResponse, error) {
	actions, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	var actionResponses []dto.ActionResponse
	for _, a := range actions {
		actionResponses = append(actionResponses, dto.ActionResponse{
			ID:        a.ID,
			Name:      a.Name,
			IsActive:  a.IsActive,
			HexCode:   a.HexCode,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		})
	}

	return actionResponses, nil
}

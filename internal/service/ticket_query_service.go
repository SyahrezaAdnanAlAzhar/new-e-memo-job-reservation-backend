package service

import (
	"context"
	"database/sql"
	"errors"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/repository"
)

type TicketQueryService struct {
	ticketRepo            *repository.TicketRepository
	trackStatusTicketRepo *repository.TrackStatusTicketRepository
	ticketActionLogRepo   *repository.TicketActionLogRepository
}

func NewTicketQueryService(ticketRepo *repository.TicketRepository, trackStatusTicketRepo *repository.TrackStatusTicketRepository, ticketActionLogRepo *repository.TicketActionLogRepository) *TicketQueryService {
	return &TicketQueryService{
		ticketRepo:            ticketRepo,
		trackStatusTicketRepo: trackStatusTicketRepo,
		ticketActionLogRepo:   ticketActionLogRepo,
	}
}

// GET ALL
func (s *TicketQueryService) GetAllTickets(filters dto.TicketFilter) ([]dto.TicketDetailResponse, error) {
	return s.ticketRepo.FindAll(filters)
}

// GET BY ID
func (s *TicketQueryService) GetTicketByID(id int) (*dto.TicketDetailResponse, error) {
	return s.ticketRepo.FindByID(id)
}

func (s *TicketQueryService) GetTicketSummary(filters dto.TicketSummaryFilter) ([]dto.TicketSummaryResponse, error) {
	return s.ticketRepo.GetTicketSummary(filters)
}

func (s *TicketQueryService) GetOldestTicket() (*dto.OldestTicketResponse, error) {
	return s.ticketRepo.FindOldestTicket()
}

func (s *TicketQueryService) GetLastRejectionDetail(ctx context.Context, ticketID int) (*dto.RejectionDetailResponse, error) {
	_, currentStatusName, err := s.trackStatusTicketRepo.GetCurrentStatusByTicketID(ctx, ticketID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("ticket not found")
		}
		return nil, err
	}

	if currentStatusName == "Ditolak" || currentStatusName == "Laporan Ditolak" {
		return s.ticketActionLogRepo.FindLastRejectionByTicketID(ctx, ticketID)
	}

	return nil, nil
}

package service

import (
	"context"
	"path/filepath"
	"strings"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"
)

type FileService struct {
	ticketRepo *repository.TicketRepository
	jobRepo    *repository.JobRepository
}

func NewFileService(ticketRepo *repository.TicketRepository, jobRepo *repository.JobRepository) *FileService {
	return &FileService{
		ticketRepo: ticketRepo,
		jobRepo:    jobRepo,
	}
}

func (s *FileService) GetAllFilesByTicketID(ctx context.Context, ticketID int) (*dto.AllFilesResponse, error) {
	supportFilesMetadata, _, err := s.ticketRepo.GetSupportFilesByTicketID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	reportFilesMetadata, _, err := s.jobRepo.GetReportFilesByTicketID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	response := &dto.AllFilesResponse{
		SupportFiles: formatFileMetadataToResponse(supportFilesMetadata),
		ReportFiles:  formatFileMetadataToResponse(reportFilesMetadata),
	}

	return response, nil
}

func formatFileMetadataToResponse(metadata []model.FileMetadata) []dto.FileResponse {
	if len(metadata) == 0 {
		return []dto.FileResponse{}
	}

	responses := make([]dto.FileResponse, len(metadata))
	for i, m := range metadata {
		responses[i] = dto.FileResponse{
			FileName:   m.FileName,
			FilePath:   m.FilePath,
			FileSize:   m.FileSize,
			FileType:   determineFileType(m.FilePath),
			UploadedAt: m.UploadedAt,
		}
	}
	return responses
}

func determineFileType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp":
		return "image"
	case ".pdf":
		return "pdf"
	case ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt":
		return "document"
	case ".mp4", ".mov", ".avi", ".mkv":
		return "video"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "archive"
	default:
		return "unknown"
	}
}

package dto

import "time"

type FileResponse struct {
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	FileSize   int64     `json:"file_size"`
	FileType   string    `json:"file_type"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type AllFilesResponse struct {
	SupportFiles []FileResponse `json:"support_files"`
	ReportFiles  []FileResponse `json:"report_files"`
}

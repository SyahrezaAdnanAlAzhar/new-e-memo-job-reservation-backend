package filehandler

import (
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"e-memo-job-reservation-api/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SaveFiles(c *gin.Context, files []*multipart.FileHeader) ([]model.FileMetadata, error) {
	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = "./uploads"
	}
	if err := os.MkdirAll(storagePath, os.ModePerm); err != nil {
		return nil, err
	}

	var savedFilesMetadata []model.FileMetadata
	for _, file := range files {
		extension := filepath.Ext(file.Filename)
		newFileName := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), uuid.New().String(), extension)
		filePath := filepath.Join(storagePath, newFileName)

		if err := c.SaveUploadedFile(file, filePath); err != nil {
			log.Printf("Error saving file %s, rolling back saved files...", file.Filename)
			for _, metadata := range savedFilesMetadata {
				os.Remove(metadata.FilePath)
			}
			return nil, err
		}

		metadata := model.FileMetadata{
			FileName:   file.Filename,
			FilePath:   filePath,
			FileSize:   file.Size,
			MimeType:   file.Header.Get("Content-Type"),
			UploadedAt: time.Now(),
		}
		savedFilesMetadata = append(savedFilesMetadata, metadata)
	}

	return savedFilesMetadata, nil
}

package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type FileMetadata struct {
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	FileSize   int64     `json:"file_size"`
	MimeType   string    `json:"mime_type"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func (fm FileMetadata) Value() (driver.Value, error) {
	return json.Marshal(fm)
}

func (fm *FileMetadata) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &fm)
}

package dto

import "time"

type FileUploadRequest struct {
	TaskID   uint   `validate:"required"`
	FileName string `validate:"required,max=255"`
	FileData []byte `validate:"required"`
	MimeType string `validate:"required,max=100"`
}

type FileResponse struct {
	ID        uint      `json:"id"`
	TaskID    uint      `json:"task_id"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

type FileCommon struct {
	ID       uint   `json:"id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
}

package dto

import "time"

type ExportRequest struct {
	GroupID  uint       `validate:"required"`
	Format   string     `validate:"required,oneof=excel pdf"`
	FromDate *time.Time `validate:"omitempty"`
	ToDate   *time.Time `validate:"omitempty"`
	Status   *string    `validate:"omitempty,oneof=pending completed overdue"`
}

type ExportResponse struct {
	FileName string `json:"file_name"`
	FileData []byte `json:"file_data"`
	MimeType string `json:"mime_type"`
}

type ExportTaskData struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	Creator     string    `json:"creator"`
	Assignee    string    `json:"assignee,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

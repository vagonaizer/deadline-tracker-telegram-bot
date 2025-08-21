package dto

import "time"

type CreateNotificationRequest struct {
	UserID uint   `validate:"required"`
	TaskID uint   `validate:"required"`
	Type   string `validate:"required,oneof=1day 3days 1week"`
}

type NotificationResponse struct {
	ID        uint       `json:"id"`
	User      UserCommon `json:"user"`
	Task      TaskCommon `json:"task"`
	Type      string     `json:"type"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type NotificationCommon struct {
	ID     uint       `json:"id"`
	Type   string     `json:"type"`
	SentAt *time.Time `json:"sent_at,omitempty"`
}

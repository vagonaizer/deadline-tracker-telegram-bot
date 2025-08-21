package dto

import "time"

type UserRequest struct {
	TelegramID int64  `validate:"required"`
	Username   string `validate:"max=100"`
	FirstName  string `validate:"max=100"`
	LastName   string `validate:"max=100"`
}

type UserResponse struct {
	ID         uint      `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserCommon struct {
	ID         uint   `json:"id"`
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username"`
	FirstName  string `json:"first_name"`
}

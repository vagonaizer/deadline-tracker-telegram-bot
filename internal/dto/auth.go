package dto

type AuthRequest struct {
	TelegramID int64 `validate:"required"`
	GroupID    uint  `validate:"required"`
}

type AuthResponse struct {
	Authorized bool        `json:"authorized"`
	User       UserCommon  `json:"user"`
	Group      GroupCommon `json:"group"`
	Role       string      `json:"role"`
}

type RoleChangeRequest struct {
	UserID  uint   `validate:"required"`
	GroupID uint   `validate:"required"`
	NewRole string `validate:"required,oneof=admin moderator member"`
}

package dto

import "time"

type CreateGroupRequest struct {
	Name     string `validate:"required,min=1,max=200"`
	Login    string `validate:"required,min=3,max=50,alphanum"`
	Password string `validate:"required,min=6,max=100"`
}

type ConnectGroupRequest struct {
	Login    string `validate:"required"`
	Password string `validate:"required"`
}

type GroupResponse struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Login     string     `json:"login"`
	Creator   UserCommon `json:"creator"`
	CreatedAt time.Time  `json:"created_at"`
}

type GroupCommon struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Login string `json:"login"`
}

type GroupMemberResponse struct {
	User     UserCommon `json:"user"`
	Role     string     `json:"role"`
	JoinedAt time.Time  `json:"joined_at"`
}

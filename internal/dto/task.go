package dto

import "time"

type CreateTaskRequest struct {
	Title       string    `validate:"required,min=1,max=200"`
	Description string    `validate:"max=1000"`
	Deadline    time.Time `validate:"required"`
	AssignedTo  *uint     `validate:"omitempty"`
}

type UpdateTaskRequest struct {
	Title       *string    `validate:"omitempty,min=1,max=200"`
	Description *string    `validate:"omitempty,max=1000"`
	Deadline    *time.Time `validate:"omitempty"`
	Priority    *string    `validate:"omitempty,oneof=low normal high critical"`
	Status      *string    `validate:"omitempty,oneof=pending completed overdue"`
	AssignedTo  *uint      `validate:"omitempty"`
}

type TaskResponse struct {
	ID          uint         `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Deadline    time.Time    `json:"deadline"`
	Priority    string       `json:"priority"`
	Status      string       `json:"status"`
	Group       GroupCommon  `json:"group"`
	Creator     UserCommon   `json:"creator"`
	Assignee    *UserCommon  `json:"assignee,omitempty"`
	Files       []FileCommon `json:"files,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
}

type TaskCommon struct {
	ID       uint      `json:"id"`
	Title    string    `json:"title"`
	Deadline time.Time `json:"deadline"`
	Priority string    `json:"priority"`
	Status   string    `json:"status"`
}

type TaskListResponse struct {
	Tasks []TaskResponse `json:"tasks"`
	Total int            `json:"total"`
}

type TaskFilterRequest struct {
	Status     *string    `validate:"omitempty,oneof=pending completed overdue"`
	Priority   *string    `validate:"omitempty,oneof=low normal high critical"`
	AssignedTo *uint      `validate:"omitempty"`
	FromDate   *time.Time `validate:"omitempty"`
	ToDate     *time.Time `validate:"omitempty"`
}

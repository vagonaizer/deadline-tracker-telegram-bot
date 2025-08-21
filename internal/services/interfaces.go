package services

import (
	"deadline-bot/internal/dto"
)

type UserService interface {
	CreateUser(req *dto.UserRequest) (*dto.UserResponse, error)
	GetUserByTelegramID(telegramID int64) (*dto.UserResponse, error)
	GetUserByID(id uint) (*dto.UserResponse, error)
	UpdateUser(id uint, req *dto.UserRequest) (*dto.UserResponse, error)
	DeleteUser(id uint) error
}

type GroupService interface {
	CreateGroup(creatorTelegramID int64, req *dto.CreateGroupRequest) (*dto.GroupResponse, error)
	ConnectToGroup(telegramID int64, req *dto.ConnectGroupRequest) (*dto.GroupResponse, error)
	GetGroupByID(id uint) (*dto.GroupResponse, error)
	GetUserGroups(telegramID int64) ([]dto.GroupResponse, error)
	GetGroupMembers(groupID uint) ([]dto.GroupMemberResponse, error)
	LeaveGroup(telegramID int64, groupID uint) error
	ChangeUserRole(adminTelegramID int64, req *dto.RoleChangeRequest) error
	DeleteGroup(adminTelegramID int64, groupID uint) error
}

type TaskService interface {
	CreateTask(creatorTelegramID int64, groupID uint, req *dto.CreateTaskRequest) (*dto.TaskResponse, error)
	GetTaskByID(id uint) (*dto.TaskResponse, error)
	GetGroupTasks(groupID uint, filter *dto.TaskFilterRequest) (*dto.TaskListResponse, error)
	GetUserTasks(telegramID int64, filter *dto.TaskFilterRequest) (*dto.TaskListResponse, error)
	UpdateTask(editorTelegramID int64, taskID uint, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error)
	CompleteTask(telegramID int64, taskID uint) error
	DeleteTask(telegramID int64, taskID uint) error
	UpdateTaskPriorities() error
	GetOverdueTasks(groupID uint) ([]dto.TaskResponse, error)
}

type AuthService interface {
	AuthorizeUser(req *dto.AuthRequest) (*dto.AuthResponse, error)
	CheckPermission(telegramID int64, groupID uint, requiredRole string) error
	GetUserRole(telegramID int64, groupID uint) (string, error)
	IsUserInGroup(telegramID int64, groupID uint) (bool, error)
}

type FileService interface {
	UploadFile(telegramID int64, req *dto.FileUploadRequest) (*dto.FileResponse, error)
	GetTaskFiles(taskID uint) ([]dto.FileResponse, error)
	GetFileByID(id uint) (*dto.FileResponse, error)
	DeleteFile(telegramID int64, fileID uint) error
	GetFilePath(id uint) (string, error)
	ValidateFile(fileName string, fileSize int64, mimeType string) error
}

type NotificationSender interface {
	SendNotification(userTelegramID int64, message string) error
}

// Обновленный NotificationService с возможностью установки sender'а
type NotificationService interface {
	CreateNotificationsForTask(taskID uint) error
	ProcessPendingNotifications() error
	GetUserNotifications(telegramID int64) ([]dto.NotificationResponse, error)
	GetPendingNotifications() ([]dto.NotificationResponse, error)
	MarkNotificationSent(id uint) error
	DeleteTaskNotifications(taskID uint) error
	CheckAndSendNotifications() error
	CreateTestNotifications(telegramID int64) error

	// Метод для установки отправителя уведомлений
	SetNotificationSender(sender NotificationSender)
}

type ExportService interface {
	ExportToExcel(telegramID int64, req *dto.ExportRequest) (*dto.ExportResponse, error)
	ExportToPDF(telegramID int64, req *dto.ExportRequest) (*dto.ExportResponse, error)
	PrepareExportData(groupID uint, filter *dto.TaskFilterRequest) ([]dto.ExportTaskData, error)
}

type SchedulerService interface {
	StartScheduler() error
	StopScheduler() error
	CheckDeadlines() error
	UpdateTaskStatuses() error
	SendReminders() error
}

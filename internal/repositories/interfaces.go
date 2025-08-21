package repositories

import (
	"deadline-bot/internal/models"
	"time"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByTelegramID(telegramID int64) (*models.User, error)
	GetByID(id uint) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
}

type GroupRepository interface {
	Create(group *models.Group) error
	GetByID(id uint) (*models.Group, error)
	GetByLogin(login string) (*models.Group, error)
	GetByUserID(userID uint) ([]models.Group, error)
	Update(group *models.Group) error
	Delete(id uint) error
}

type UserGroupRepository interface {
	Create(userGroup *models.UserGroup) error
	GetByUserAndGroup(userID, groupID uint) (*models.UserGroup, error)
	GetUsersByGroup(groupID uint) ([]models.UserGroup, error)
	GetGroupsByUser(userID uint) ([]models.UserGroup, error)
	UpdateRole(userID, groupID uint, role string) error
	Delete(userID, groupID uint) error
	IsUserInGroup(userID, groupID uint) (bool, error)
	GetUserRole(userID, groupID uint) (string, error)
}

type TaskRepository interface {
	Create(task *models.Task) error
	GetByID(id uint) (*models.Task, error)
	GetByGroupID(groupID uint) ([]models.Task, error)
	GetByUserID(userID uint) ([]models.Task, error)
	GetByStatus(groupID uint, status string) ([]models.Task, error)
	GetByPriority(groupID uint, priority string) ([]models.Task, error)
	GetByDeadlineRange(groupID uint, from, to time.Time) ([]models.Task, error)
	GetOverdueTasks(groupID uint) ([]models.Task, error)
	Update(task *models.Task) error
	UpdateStatus(id uint, status string) error
	UpdatePriority(id uint, priority string) error
	Delete(id uint) error
}

type FileRepository interface {
	Create(file *models.File) error
	GetByID(id uint) (*models.File, error)
	GetByTaskID(taskID uint) ([]models.File, error)
	Delete(id uint) error
	DeleteByTaskID(taskID uint) error
}

type NotificationRepository interface {
	Create(notification *models.Notification) error
	GetByID(id uint) (*models.Notification, error)
	GetPendingByType(notificationType string) ([]models.Notification, error)
	GetByUserAndTask(userID, taskID uint) ([]models.Notification, error)
	MarkAsSent(id uint) error
	Delete(id uint) error
	DeleteByTaskID(taskID uint) error
	GetPendingByUserID(userID uint) ([]models.Notification, error)
	GetReadyToSend(now time.Time) ([]models.Notification, error)
}

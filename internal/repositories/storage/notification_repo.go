package storage

import (
	"deadline-bot/internal/models"
	"deadline-bot/internal/repositories"
	"time"

	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) repositories.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetByID(id uint) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.Preload("User").
		Preload("Task").
		First(&notification, id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) GetPendingByType(notificationType string) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("type = ? AND sent_at IS NULL", notificationType).
		Preload("User").
		Preload("Task").
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) GetByUserAndTask(userID, taskID uint) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("user_id = ? AND task_id = ?", userID, taskID).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) MarkAsSent(id uint) error {
	now := time.Now()
	return r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Update("sent_at", &now).Error
}

func (r *notificationRepository) Delete(id uint) error {
	return r.db.Delete(&models.Notification{}, id).Error
}

func (r *notificationRepository) DeleteByTaskID(taskID uint) error {
	return r.db.Where("task_id = ?", taskID).Delete(&models.Notification{}).Error
}

func (r *notificationRepository) GetReadyToSend(now time.Time) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("scheduled_time <= ? AND sent_at IS NULL", now).
		Preload("Task").
		Preload("User").
		Find(&notifications).Error

	return notifications, err
}

func (r *notificationRepository) GetPendingByUserID(userID uint) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("user_id = ? AND sent_at IS NULL", userID).
		Preload("Task").
		Preload("User").
		Order("scheduled_time ASC").
		Find(&notifications).Error

	return notifications, err
}

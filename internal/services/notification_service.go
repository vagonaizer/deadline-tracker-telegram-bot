package services

import (
	"deadline-bot/internal/config"
	"deadline-bot/internal/dto"
	"deadline-bot/internal/models"
	"deadline-bot/internal/repositories"
	"deadline-bot/pkg/utils"

	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type notificationService struct {
	notificationRepo repositories.NotificationRepository
	taskRepo         repositories.TaskRepository
	userRepo         repositories.UserRepository
	userGroupRepo    repositories.UserGroupRepository
	config           *config.NotificationsConfig
	sender           NotificationSender
}

func NewNotificationService(
	notificationRepo repositories.NotificationRepository,
	taskRepo repositories.TaskRepository,
	userRepo repositories.UserRepository,
	userGroupRepo repositories.UserGroupRepository,
	config *config.NotificationsConfig,
) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		taskRepo:         taskRepo,
		userRepo:         userRepo,
		userGroupRepo:    userGroupRepo,
		config:           config,
	}
}

func (s *notificationService) SetNotificationSender(sender NotificationSender) {
	s.sender = sender
	log.Println("✅ Notification sender установлен")
}

// ПРОСТАЯ ЛОГИКА: Создаем уведомления с точным временем отправки
func (s *notificationService) CreateNotificationsForTask(taskID uint) error {
	if !s.config.EnableReminders {
		log.Println("DEBUG: Reminders disabled")
		return nil
	}

	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	userGroups, err := s.userGroupRepo.GetUsersByGroup(task.GroupID)
	if err != nil {
		return fmt.Errorf("failed to get group users: %w", err)
	}

	created := 0
	now := time.Now()

	for _, userGroup := range userGroups {
		for _, days := range s.config.DefaultReminders {
			// Вычисляем ТОЧНОЕ время отправки уведомления
			scheduledTime := task.Deadline.AddDate(0, 0, -days)

			// Если время уже прошло, но дедлайн не наступил - отправляем сейчас
			if scheduledTime.Before(now) && task.Deadline.After(now) {
				scheduledTime = now.Add(1 * time.Minute) // отправим через минуту
			}

			// Если дедлайн уже прошел - пропускаем
			if task.Deadline.Before(now) {
				continue
			}

			var notificationType string
			switch days {
			case 1:
				notificationType = models.NotificationTypeOneDay
			case 3:
				notificationType = models.NotificationTypeThreeDays
			case 7:
				notificationType = models.NotificationTypeOneWeek
			default:
				continue
			}

			notification := &models.Notification{
				UserID:        userGroup.UserID,
				TaskID:        taskID,
				Type:          notificationType,
				ScheduledTime: &scheduledTime, // НОВОЕ ПОЛЕ!
			}

			if err := s.notificationRepo.Create(notification); err != nil {
				log.Printf("❌ Failed to create notification: %v", err)
				continue
			}

			created++
			log.Printf("✅ Created %s notification for user %d, scheduled for %v",
				notificationType, userGroup.UserID, scheduledTime)
		}
	}

	log.Printf("📋 Created %d notifications for task %d", created, taskID)
	return nil
}

// ПРОСТАЯ ЛОГИКА: Ищем уведомления, где scheduled_time <= now
func (s *notificationService) ProcessPendingNotifications() error {
	if !s.config.EnableReminders {
		return nil
	}

	if s.sender == nil {
		log.Println("⚠️ Notification sender not set")
		return nil
	}

	log.Println("🔍 Processing pending notifications...")

	// Ищем все уведомления, которые пора отправить
	notifications, err := s.notificationRepo.GetReadyToSend(time.Now())
	if err != nil {
		return fmt.Errorf("failed to get ready notifications: %w", err)
	}

	log.Printf("🔍 Found %d notifications ready to send", len(notifications))

	sent := 0
	for _, notification := range notifications {
		log.Printf("📤 Sending notification %d to user %d", notification.ID, notification.UserID)

		if err := s.sendNotification(&notification); err != nil {
			log.Printf("❌ Failed to send notification %d: %v", notification.ID, err)
			continue
		}

		if err := s.notificationRepo.MarkAsSent(notification.ID); err != nil {
			log.Printf("❌ Failed to mark notification %d as sent: %v", notification.ID, err)
			continue
		}

		sent++
		log.Printf("✅ Successfully sent notification %d", notification.ID)
	}

	log.Printf("📤 Sent %d notifications", sent)
	return nil
}

func (s *notificationService) sendNotification(notification *models.Notification) error {
	if s.sender == nil {
		return fmt.Errorf("notification sender not set")
	}

	user, err := s.userRepo.GetByID(notification.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	task, err := s.taskRepo.GetByID(notification.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	var typeText string
	switch notification.Type {
	case models.NotificationTypeOneDay:
		typeText = "завтра"
	case models.NotificationTypeThreeDays:
		typeText = "через 3 дня"
	case models.NotificationTypeOneWeek:
		typeText = "через неделю"
	default:
		typeText = "скоро"
	}

	message := fmt.Sprintf("🔔 Напоминание о дедлайне!\n\n"+
		"📋 Задача: %s\n"+
		"⏰ Дедлайн %s (%s)\n"+
		"📝 Описание: %s",
		task.Title,
		typeText,
		utils.FormatDeadline(task.Deadline),
		task.Description)

	return s.sender.SendNotification(user.TelegramID, message)
}

// Остальные методы остаются как есть...
func (s *notificationService) GetUserNotifications(telegramID int64) ([]dto.NotificationResponse, error) {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	notifications, err := s.notificationRepo.GetPendingByUserID(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	responses := make([]dto.NotificationResponse, len(notifications))
	for i, notification := range notifications {
		responses[i] = *utils.NotificationToResponse(&notification)
	}

	return responses, nil
}

func (s *notificationService) GetPendingNotifications() ([]dto.NotificationResponse, error) {
	if !s.config.EnableReminders {
		return nil, nil
	}

	notifications, err := s.notificationRepo.GetReadyToSend(time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get ready notifications: %w", err)
	}

	responses := make([]dto.NotificationResponse, len(notifications))
	for i, notification := range notifications {
		responses[i] = *utils.NotificationToResponse(&notification)
	}

	return responses, nil
}

func (s *notificationService) MarkNotificationSent(id uint) error {
	return s.notificationRepo.MarkAsSent(id)
}

func (s *notificationService) DeleteTaskNotifications(taskID uint) error {
	return s.notificationRepo.DeleteByTaskID(taskID)
}

func (s *notificationService) CheckAndSendNotifications() error {
	return s.ProcessPendingNotifications()
}

func (s *notificationService) CreateTestNotifications(telegramID int64) error {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		return fmt.Errorf("пользователь не найден: %w", err)
	}

	// Создаем тестовое уведомление на отправку через 30 секунд
	scheduledTime := time.Now().Add(30 * time.Second)

	testNotification := &models.Notification{
		UserID:        user.ID,
		TaskID:        1, // любая существующая задача
		Type:          models.NotificationTypeOneDay,
		ScheduledTime: &scheduledTime,
	}

	if err := s.notificationRepo.Create(testNotification); err != nil {
		return fmt.Errorf("failed to create test notification: %w", err)
	}

	log.Printf("✅ Created test notification for user %d, scheduled for %v", user.ID, scheduledTime)
	return nil
}

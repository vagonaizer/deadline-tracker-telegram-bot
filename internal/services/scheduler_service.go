package services

import (
	"deadline-bot/internal/config"
	"deadline-bot/internal/models"
	"deadline-bot/internal/repositories"
	"deadline-bot/pkg/utils"

	"fmt"
	"log"
	"time"
)

type schedulerService struct {
	taskRepo            repositories.TaskRepository
	groupRepo           repositories.GroupRepository
	notificationService NotificationService
	config              *config.NotificationsConfig
	ticker              *time.Ticker
	stopChan            chan bool
	running             bool
}

func NewSchedulerService(
	taskRepo repositories.TaskRepository,
	groupRepo repositories.GroupRepository,
	notificationService NotificationService,
	config *config.NotificationsConfig,
) SchedulerService {
	return &schedulerService{
		taskRepo:            taskRepo,
		groupRepo:           groupRepo,
		notificationService: notificationService,
		config:              config,
		stopChan:            make(chan bool),
	}
}

func (s *schedulerService) StartScheduler() error {
	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	if !s.config.EnableReminders {
		log.Println("❌ Notifications disabled, scheduler not started")
		return nil
	}

	// Отладочный вывод
	log.Printf("DEBUG: CheckIntervalHours from config: %f", s.config.CheckIntervalHours)

	// Проверяем на неположительное значение
	if s.config.CheckIntervalHours <= 0 {
		log.Printf("❌ Invalid interval: %f, using default 0.1 hours", s.config.CheckIntervalHours)
		s.config.CheckIntervalHours = 0.1 // принудительно устанавливаем 6 минут
	}

	// Поддерживаем дробные часы для тестирования
	intervalDuration := time.Duration(s.config.CheckIntervalHours * float64(time.Hour))
	log.Printf("DEBUG: Calculated interval duration: %v", intervalDuration)

	s.ticker = time.NewTicker(intervalDuration)
	s.running = true

	go s.run()

	log.Printf("✅ Scheduler started with interval: %v (%.1f hours)", intervalDuration, s.config.CheckIntervalHours)

	// Запускаем первую проверку сразу
	go func() {
		log.Println("🚀 Running initial notification check...")
		if err := s.CheckDeadlines(); err != nil {
			log.Printf("❌ Error in initial notification check: %v", err)
		}
	}()

	return nil
}

func (s *schedulerService) StopScheduler() error {
	if !s.running {
		return fmt.Errorf("scheduler is not running")
	}

	s.ticker.Stop()
	s.stopChan <- true
	s.running = false

	log.Println("🛑 Scheduler stopped")
	return nil
}

func (s *schedulerService) CheckDeadlines() error {
	log.Println("🔍 Checking deadlines and processing notifications...")

	if err := s.UpdateTaskStatuses(); err != nil {
		log.Printf("❌ Failed to update task statuses: %v", err)
		return err
	}

	if err := s.SendReminders(); err != nil {
		log.Printf("❌ Failed to send reminders: %v", err)
		return err
	}

	log.Println("✅ Deadline check completed")
	return nil
}

func (s *schedulerService) UpdateTaskStatuses() error {
	log.Println("📊 Updating task statuses...")

	groups, err := s.getAllGroups()
	if err != nil {
		return fmt.Errorf("failed to get groups: %w", err)
	}

	totalUpdated := 0

	for _, group := range groups {
		overdueTasks, err := s.taskRepo.GetOverdueTasks(group.ID)
		if err != nil {
			log.Printf("❌ Failed to get overdue tasks for group %d: %v", group.ID, err)
			continue
		}

		for _, task := range overdueTasks {
			if task.Status == models.StatusPending {
				if err := s.taskRepo.UpdateStatus(task.ID, models.StatusOverdue); err != nil {
					log.Printf("❌ Failed to update task %d status to overdue: %v", task.ID, err)
				} else {
					totalUpdated++
					log.Printf("📋 Task %d marked as overdue", task.ID)
				}
			}
		}

		allTasks, err := s.taskRepo.GetByGroupID(group.ID)
		if err != nil {
			log.Printf("❌ Failed to get tasks for group %d: %v", group.ID, err)
			continue
		}

		for _, task := range allTasks {
			if task.Status != models.StatusCompleted && task.Status != models.StatusOverdue {
				newPriority := utils.GetPriorityByDeadline(task.Deadline)
				if task.Priority != newPriority {
					if err := s.taskRepo.UpdatePriority(task.ID, newPriority); err != nil {
						log.Printf("❌ Failed to update task %d priority: %v", task.ID, err)
					} else {
						log.Printf("🎯 Task %d priority updated to %s", task.ID, newPriority)
					}
				}
			}
		}
	}

	log.Printf("📊 Updated %d task statuses", totalUpdated)
	return nil
}

func (s *schedulerService) SendReminders() error {
	log.Println("🔔 Processing pending notifications...")

	err := s.notificationService.ProcessPendingNotifications()
	if err != nil {
		log.Printf("❌ Error processing notifications: %v", err)
		return err
	}

	log.Println("✅ Notifications processed successfully")
	return nil
}

func (s *schedulerService) run() {
	log.Println("🔄 Scheduler main loop started")

	for {
		select {
		case <-s.ticker.C:
			log.Println("⏰ Scheduled check triggered")
			if err := s.CheckDeadlines(); err != nil {
				log.Printf("❌ Error in scheduled deadline check: %v", err)
			}
		case <-s.stopChan:
			log.Println("🛑 Scheduler main loop stopped")
			return
		}
	}
}

func (s *schedulerService) getAllGroups() ([]models.Group, error) {
	// Заглушка - в реальной системе здесь должен быть вызов groupRepo
	var groups []models.Group
	// TODO: реализовать получение всех групп
	return groups, nil
}

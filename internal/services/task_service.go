package services

import (
	"deadline-bot/internal/dto"
	"deadline-bot/internal/models"
	"deadline-bot/internal/repositories"
	"deadline-bot/pkg/utils"
	"fmt"
	"log"

	"gorm.io/gorm"
)

type taskService struct {
	userRepo            repositories.UserRepository
	taskRepo            repositories.TaskRepository
	userGroupRepo       repositories.UserGroupRepository
	authService         AuthService
	notificationService NotificationService
}

func NewTaskService(
	userRepo repositories.UserRepository,
	taskRepo repositories.TaskRepository,
	userGroupRepo repositories.UserGroupRepository,
	authService AuthService,
	notificationService NotificationService,
) TaskService {
	return &taskService{
		userRepo:            userRepo,
		taskRepo:            taskRepo,
		userGroupRepo:       userGroupRepo,
		authService:         authService,
		notificationService: notificationService,
	}
}

func (s *taskService) CreateTask(creatorTelegramID int64, groupID uint, req *dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	if err := s.authService.CheckPermission(creatorTelegramID, groupID, models.RoleMember); err != nil {
		return nil, err
	}

	if err := utils.ValidateTaskTitle(req.Title); err != nil {
		return nil, err
	}

	if err := utils.ValidateTaskDescription(req.Description); err != nil {
		return nil, err
	}

	if err := utils.ValidateDeadline(req.Deadline); err != nil {
		return nil, err
	}

	creator, err := s.userRepo.GetByTelegramID(creatorTelegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}

	if req.AssignedTo != nil {
		isInGroup, err := s.userGroupRepo.IsUserInGroup(*req.AssignedTo, groupID)
		if err != nil {
			return nil, fmt.Errorf("failed to check assignee membership: %w", err)
		}
		if !isInGroup {
			return nil, fmt.Errorf("назначенный пользователь не состоит в группе")
		}
	}

	task := utils.CreateTaskRequestToModel(req, groupID, creator.ID)

	if err := s.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	log.Printf("🆕 Task created with ID: %d", task.ID)

	// Создаём уведомления для задачи
	log.Printf("🔔 Creating notifications for task %d...", task.ID)
	if err := s.notificationService.CreateNotificationsForTask(task.ID); err != nil {
		log.Printf("❌ Failed to create notifications for task %d: %v", task.ID, err)
	} else {
		log.Printf("✅ Notifications created for task %d", task.ID)
	}

	createdTask, err := s.taskRepo.GetByID(task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created task: %w", err)
	}

	return utils.TaskToResponse(createdTask), nil
}

func (s *taskService) GetTaskByID(id uint) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("задача не найдена")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return utils.TaskToResponse(task), nil
}

func (s *taskService) GetGroupTasks(groupID uint, filter *dto.TaskFilterRequest) (*dto.TaskListResponse, error) {
	var tasks []models.Task
	var err error

	if filter == nil {
		tasks, err = s.taskRepo.GetByGroupID(groupID)
	} else {
		tasks, err = s.getFilteredTasks(groupID, filter)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	responses := make([]dto.TaskResponse, len(tasks))
	for i, task := range tasks {
		responses[i] = *utils.TaskToResponse(&task)
	}

	return &dto.TaskListResponse{
		Tasks: responses,
		Total: len(responses),
	}, nil
}

func (s *taskService) GetUserTasks(telegramID int64, filter *dto.TaskFilterRequest) (*dto.TaskListResponse, error) {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	tasks, err := s.taskRepo.GetByUserID(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tasks: %w", err)
	}

	filteredTasks := s.applyClientFilter(tasks, filter)

	responses := make([]dto.TaskResponse, len(filteredTasks))
	for i, task := range filteredTasks {
		responses[i] = *utils.TaskToResponse(&task)
	}

	return &dto.TaskListResponse{
		Tasks: responses,
		Total: len(responses),
	}, nil
}

func (s *taskService) UpdateTask(editorTelegramID int64, taskID uint, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("задача не найдена")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	editor, err := s.userRepo.GetByTelegramID(editorTelegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get editor: %w", err)
	}

	canEdit, err := s.canUserEditTask(editor.ID, task)
	if err != nil {
		return nil, err
	}

	if !canEdit {
		return nil, fmt.Errorf("недостаточно прав для редактирования задачи")
	}

	if err := s.applyTaskUpdates(task, req); err != nil {
		return nil, err
	}

	if err := s.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	updatedTask, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated task: %w", err)
	}

	return utils.TaskToResponse(updatedTask), nil
}

func (s *taskService) CompleteTask(telegramID int64, taskID uint) error {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("задача не найдена")
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	if err := s.authService.CheckPermission(telegramID, task.GroupID, models.RoleMember); err != nil {
		return err
	}

	if err := s.taskRepo.UpdateStatus(taskID, models.StatusCompleted); err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	return nil
}

func (s *taskService) DeleteTask(telegramID int64, taskID uint) error {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("задача не найдена")
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	canDelete, err := s.canUserDeleteTask(user.ID, task)
	if err != nil {
		return err
	}

	if !canDelete {
		return fmt.Errorf("недостаточно прав для удаления задачи")
	}

	if err := s.taskRepo.Delete(taskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

func (s *taskService) UpdateTaskPriorities() error {
	// Этот метод будет вызываться планировщиком для обновления приоритетов всех активных задач
	// Пока оставим пустую реализацию, так как логика будет в SchedulerService
	return nil
}

func (s *taskService) GetOverdueTasks(groupID uint) ([]dto.TaskResponse, error) {
	tasks, err := s.taskRepo.GetOverdueTasks(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue tasks: %w", err)
	}

	responses := make([]dto.TaskResponse, len(tasks))
	for i, task := range tasks {
		responses[i] = *utils.TaskToResponse(&task)
	}

	return responses, nil
}

func (s *taskService) getFilteredTasks(groupID uint, filter *dto.TaskFilterRequest) ([]models.Task, error) {
	if filter.Status != nil {
		return s.taskRepo.GetByStatus(groupID, *filter.Status)
	}

	if filter.Priority != nil {
		return s.taskRepo.GetByPriority(groupID, *filter.Priority)
	}

	if filter.FromDate != nil && filter.ToDate != nil {
		return s.taskRepo.GetByDeadlineRange(groupID, *filter.FromDate, *filter.ToDate)
	}

	return s.taskRepo.GetByGroupID(groupID)
}

func (s *taskService) applyClientFilter(tasks []models.Task, filter *dto.TaskFilterRequest) []models.Task {
	if filter == nil {
		return tasks
	}

	var filtered []models.Task

	for _, task := range tasks {
		if filter.Status != nil && task.Status != *filter.Status {
			continue
		}

		if filter.Priority != nil && task.Priority != *filter.Priority {
			continue
		}

		if filter.FromDate != nil && task.Deadline.Before(*filter.FromDate) {
			continue
		}

		if filter.ToDate != nil && task.Deadline.After(*filter.ToDate) {
			continue
		}

		filtered = append(filtered, task)
	}

	return filtered
}

func (s *taskService) canUserEditTask(userID uint, task *models.Task) (bool, error) {
	if task.CreatedBy == userID {
		return true, nil
	}

	role, err := s.userGroupRepo.GetUserRole(userID, task.GroupID)
	if err != nil {
		return false, fmt.Errorf("failed to get user role: %w", err)
	}

	return role == models.RoleAdmin || role == models.RoleModerator, nil
}

func (s *taskService) canUserDeleteTask(userID uint, task *models.Task) (bool, error) {
	if task.CreatedBy == userID {
		return true, nil
	}

	role, err := s.userGroupRepo.GetUserRole(userID, task.GroupID)
	if err != nil {
		return false, fmt.Errorf("failed to get user role: %w", err)
	}

	return role == models.RoleAdmin, nil
}

func (s *taskService) applyTaskUpdates(task *models.Task, req *dto.UpdateTaskRequest) error {
	if req.Title != nil {
		if err := utils.ValidateTaskTitle(*req.Title); err != nil {
			return err
		}
		task.Title = *req.Title
	}

	if req.Description != nil {
		if err := utils.ValidateTaskDescription(*req.Description); err != nil {
			return err
		}
		task.Description = *req.Description
	}

	if req.Deadline != nil {
		if err := utils.ValidateDeadline(*req.Deadline); err != nil {
			return err
		}
		task.Deadline = *req.Deadline
		task.Priority = utils.GetPriorityByDeadline(*req.Deadline)
	}

	if req.Priority != nil {
		if err := utils.ValidatePriority(*req.Priority); err != nil {
			return err
		}
		task.Priority = *req.Priority
	}

	if req.Status != nil {
		if err := utils.ValidateStatus(*req.Status); err != nil {
			return err
		}
		task.Status = *req.Status
	}

	if req.AssignedTo != nil {
		isInGroup, err := s.userGroupRepo.IsUserInGroup(*req.AssignedTo, task.GroupID)
		if err != nil {
			return fmt.Errorf("failed to check assignee membership: %w", err)
		}
		if !isInGroup {
			return fmt.Errorf("назначенный пользователь не состоит в группе")
		}
		task.AssignedTo = req.AssignedTo
	}

	return nil
}

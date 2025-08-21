package storage

import (
	"deadline-bot/internal/models"
	"deadline-bot/internal/repositories"
	"time"

	"gorm.io/gorm"
)

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) repositories.TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(task *models.Task) error {
	return r.db.Create(task).Error
}

func (r *taskRepository) GetByID(id uint) (*models.Task, error) {
	var task models.Task
	err := r.db.Preload("Group").
		Preload("Creator").
		Preload("Assignee").
		Preload("Files").
		First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) GetByGroupID(groupID uint) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Where("group_id = ?", groupID).
		Preload("Creator").
		Preload("Assignee").
		Preload("Files").
		Order("deadline ASC").
		Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) GetByUserID(userID uint) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Where("assigned_to = ?", userID).
		Preload("Group").
		Preload("Creator").
		Preload("Files").
		Order("deadline ASC").
		Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) GetByStatus(groupID uint, status string) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Where("group_id = ? AND status = ?", groupID, status).
		Preload("Creator").
		Preload("Assignee").
		Order("deadline ASC").
		Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) GetByPriority(groupID uint, priority string) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Where("group_id = ? AND priority = ?", groupID, priority).
		Preload("Creator").
		Preload("Assignee").
		Order("deadline ASC").
		Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) GetByDeadlineRange(groupID uint, from, to time.Time) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Where("group_id = ? AND deadline BETWEEN ? AND ?", groupID, from, to).
		Preload("Creator").
		Preload("Assignee").
		Order("deadline ASC").
		Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) GetOverdueTasks(groupID uint) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Where("group_id = ? AND deadline < ? AND status != ?",
		groupID, time.Now(), models.StatusCompleted).
		Preload("Creator").
		Preload("Assignee").
		Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) Update(task *models.Task) error {
	return r.db.Save(task).Error
}

func (r *taskRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Task{}).Where("id = ?", id).Update("status", status).Error
}

func (r *taskRepository) UpdatePriority(id uint, priority string) error {
	return r.db.Model(&models.Task{}).Where("id = ?", id).Update("priority", priority).Error
}

func (r *taskRepository) Delete(id uint) error {
	return r.db.Delete(&models.Task{}, id).Error
}

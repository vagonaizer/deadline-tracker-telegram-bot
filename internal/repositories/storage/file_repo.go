package storage

import (
	"deadline-bot/internal/models"
	"deadline-bot/internal/repositories"

	"gorm.io/gorm"
)

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) repositories.FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(file *models.File) error {
	return r.db.Create(file).Error
}

func (r *fileRepository) GetByID(id uint) (*models.File, error) {
	var file models.File
	err := r.db.Preload("Task").First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepository) GetByTaskID(taskID uint) ([]models.File, error) {
	var files []models.File
	err := r.db.Where("task_id = ?", taskID).
		Order("created_at ASC").
		Find(&files).Error
	return files, err
}

func (r *fileRepository) Delete(id uint) error {
	return r.db.Delete(&models.File{}, id).Error
}

func (r *fileRepository) DeleteByTaskID(taskID uint) error {
	return r.db.Where("task_id = ?", taskID).Delete(&models.File{}).Error
}

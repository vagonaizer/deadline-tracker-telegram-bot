package storage

import (
	"deadline-bot/internal/models"
	"deadline-bot/internal/repositories"

	"gorm.io/gorm"
)

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) repositories.GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) Create(group *models.Group) error {
	return r.db.Create(group).Error
}

func (r *groupRepository) GetByID(id uint) (*models.Group, error) {
	var group models.Group
	err := r.db.Preload("Creator").First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepository) GetByLogin(login string) (*models.Group, error) {
	var group models.Group
	err := r.db.Where("login = ?", login).Preload("Creator").First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepository) GetByUserID(userID uint) ([]models.Group, error) {
	var groups []models.Group
	err := r.db.Joins("JOIN user_groups ON user_groups.group_id = groups.id").
		Where("user_groups.user_id = ?", userID).
		Preload("Creator").
		Find(&groups).Error
	return groups, err
}

func (r *groupRepository) Update(group *models.Group) error {
	return r.db.Save(group).Error
}

func (r *groupRepository) Delete(id uint) error {
	return r.db.Delete(&models.Group{}, id).Error
}

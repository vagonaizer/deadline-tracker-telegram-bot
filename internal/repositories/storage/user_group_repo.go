package storage

import (
	"deadline-bot/internal/models"
	"deadline-bot/internal/repositories"

	"gorm.io/gorm"
)

type userGroupRepository struct {
	db *gorm.DB
}

func NewUserGroupRepository(db *gorm.DB) repositories.UserGroupRepository {
	return &userGroupRepository{db: db}
}

func (r *userGroupRepository) Create(userGroup *models.UserGroup) error {
	return r.db.Create(userGroup).Error
}

func (r *userGroupRepository) GetByUserAndGroup(userID, groupID uint) (*models.UserGroup, error) {
	var userGroup models.UserGroup
	err := r.db.Where("user_id = ? AND group_id = ?", userID, groupID).
		Preload("User").
		Preload("Group").
		First(&userGroup).Error
	if err != nil {
		return nil, err
	}
	return &userGroup, nil
}

func (r *userGroupRepository) GetUsersByGroup(groupID uint) ([]models.UserGroup, error) {
	var userGroups []models.UserGroup
	err := r.db.Where("group_id = ?", groupID).
		Preload("User").
		Find(&userGroups).Error
	return userGroups, err
}

func (r *userGroupRepository) GetGroupsByUser(userID uint) ([]models.UserGroup, error) {
	var userGroups []models.UserGroup
	err := r.db.Where("user_id = ?", userID).
		Preload("Group").
		Find(&userGroups).Error
	return userGroups, err
}

func (r *userGroupRepository) UpdateRole(userID, groupID uint, role string) error {
	return r.db.Model(&models.UserGroup{}).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		Update("role", role).Error
}

func (r *userGroupRepository) Delete(userID, groupID uint) error {
	return r.db.Where("user_id = ? AND group_id = ?", userID, groupID).
		Delete(&models.UserGroup{}).Error
}

func (r *userGroupRepository) IsUserInGroup(userID, groupID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserGroup{}).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		Count(&count).Error
	return count > 0, err
}

func (r *userGroupRepository) GetUserRole(userID, groupID uint) (string, error) {
	var userGroup models.UserGroup
	err := r.db.Select("role").
		Where("user_id = ? AND group_id = ?", userID, groupID).
		First(&userGroup).Error
	return userGroup.Role, err
}

package services

import (
	"deadline-bot/internal/dto"
	"deadline-bot/internal/models"
	"deadline-bot/internal/repositories"
	"deadline-bot/pkg/utils"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type groupService struct {
	userRepo      repositories.UserRepository
	groupRepo     repositories.GroupRepository
	userGroupRepo repositories.UserGroupRepository
	authService   AuthService
}

func NewGroupService(
	userRepo repositories.UserRepository,
	groupRepo repositories.GroupRepository,
	userGroupRepo repositories.UserGroupRepository,
	authService AuthService,
) GroupService {
	return &groupService{
		userRepo:      userRepo,
		groupRepo:     groupRepo,
		userGroupRepo: userGroupRepo,
		authService:   authService,
	}
}

func (s *groupService) CreateGroup(creatorTelegramID int64, req *dto.CreateGroupRequest) (*dto.GroupResponse, error) {
	if err := utils.ValidateGroupLogin(req.Login); err != nil {
		return nil, err
	}

	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	creator, err := s.userRepo.GetByTelegramID(creatorTelegramID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("создатель группы не найден")
		}
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}

	existingGroup, err := s.groupRepo.GetByLogin(req.Login)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing group: %w", err)
	}

	if existingGroup != nil {
		return nil, fmt.Errorf("группа с таким логином уже существует")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	group := utils.CreateGroupRequestToModel(req, creator.ID, hashedPassword)

	if err := s.groupRepo.Create(group); err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	userGroup := &models.UserGroup{
		UserID:   creator.ID,
		GroupID:  group.ID,
		Role:     models.RoleAdmin,
		JoinedAt: time.Now(),
	}

	if err := s.userGroupRepo.Create(userGroup); err != nil {
		return nil, fmt.Errorf("failed to add creator to group: %w", err)
	}

	group.Creator = *creator
	return utils.GroupToResponse(group), nil
}

func (s *groupService) ConnectToGroup(telegramID int64, req *dto.ConnectGroupRequest) (*dto.GroupResponse, error) {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	group, err := s.groupRepo.GetByLogin(req.Login)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("группа не найдена")
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	valid, err := utils.VerifyPassword(req.Password, group.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}

	if !valid {
		return nil, fmt.Errorf("неверный пароль")
	}

	isAlreadyMember, err := s.userGroupRepo.IsUserInGroup(user.ID, group.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}

	if isAlreadyMember {
		return utils.GroupToResponse(group), nil
	}

	userGroup := &models.UserGroup{
		UserID:   user.ID,
		GroupID:  group.ID,
		Role:     models.RoleMember,
		JoinedAt: time.Now(),
	}

	if err := s.userGroupRepo.Create(userGroup); err != nil {
		return nil, fmt.Errorf("failed to join group: %w", err)
	}

	return utils.GroupToResponse(group), nil
}

func (s *groupService) GetGroupByID(id uint) (*dto.GroupResponse, error) {
	group, err := s.groupRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("группа не найдена")
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	return utils.GroupToResponse(group), nil
}

func (s *groupService) GetUserGroups(telegramID int64) ([]dto.GroupResponse, error) {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	groups, err := s.groupRepo.GetByUserID(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	responses := make([]dto.GroupResponse, len(groups))
	for i, group := range groups {
		responses[i] = *utils.GroupToResponse(&group)
	}

	return responses, nil
}

func (s *groupService) GetGroupMembers(groupID uint) ([]dto.GroupMemberResponse, error) {
	userGroups, err := s.userGroupRepo.GetUsersByGroup(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}

	responses := make([]dto.GroupMemberResponse, len(userGroups))
	for i, userGroup := range userGroups {
		responses[i] = *utils.UserGroupToMemberResponse(&userGroup)
	}

	return responses, nil
}

func (s *groupService) LeaveGroup(telegramID int64, groupID uint) error {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("пользователь не найден")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	isInGroup, err := s.userGroupRepo.IsUserInGroup(user.ID, groupID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}

	if !isInGroup {
		return fmt.Errorf("пользователь не состоит в группе")
	}

	role, err := s.userGroupRepo.GetUserRole(user.ID, groupID)
	if err != nil {
		return fmt.Errorf("failed to get user role: %w", err)
	}

	if role == models.RoleAdmin {
		members, err := s.userGroupRepo.GetUsersByGroup(groupID)
		if err != nil {
			return fmt.Errorf("failed to get group members: %w", err)
		}

		if len(members) > 1 {
			return fmt.Errorf("админ не может покинуть группу пока в ней есть другие участники")
		}
	}

	if err := s.userGroupRepo.Delete(user.ID, groupID); err != nil {
		return fmt.Errorf("failed to leave group: %w", err)
	}

	return nil
}

func (s *groupService) ChangeUserRole(adminTelegramID int64, req *dto.RoleChangeRequest) error {
	if err := s.authService.CheckPermission(adminTelegramID, req.GroupID, models.RoleAdmin); err != nil {
		return err
	}

	if err := utils.ValidateRole(req.NewRole); err != nil {
		return err
	}

	isInGroup, err := s.userGroupRepo.IsUserInGroup(req.UserID, req.GroupID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}

	if !isInGroup {
		return fmt.Errorf("пользователь не состоит в группе")
	}

	if err := s.userGroupRepo.UpdateRole(req.UserID, req.GroupID, req.NewRole); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

func (s *groupService) DeleteGroup(adminTelegramID int64, groupID uint) error {
	if err := s.authService.CheckPermission(adminTelegramID, groupID, models.RoleAdmin); err != nil {
		return err
	}

	if err := s.groupRepo.Delete(groupID); err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	return nil
}

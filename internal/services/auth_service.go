package services

import (
	"deadline-bot/internal/dto"
	"deadline-bot/internal/models"
	"deadline-bot/internal/repositories"
	"deadline-bot/pkg/utils"
	"fmt"

	"gorm.io/gorm"
)

type authService struct {
	userRepo      repositories.UserRepository
	groupRepo     repositories.GroupRepository
	userGroupRepo repositories.UserGroupRepository
}

func NewAuthService(
	userRepo repositories.UserRepository,
	groupRepo repositories.GroupRepository,
	userGroupRepo repositories.UserGroupRepository,
) AuthService {
	return &authService{
		userRepo:      userRepo,
		groupRepo:     groupRepo,
		userGroupRepo: userGroupRepo,
	}
}

func (s *authService) AuthorizeUser(req *dto.AuthRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.GetByTelegramID(req.TelegramID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &dto.AuthResponse{Authorized: false}, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	group, err := s.groupRepo.GetByID(req.GroupID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &dto.AuthResponse{Authorized: false}, nil
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	isInGroup, err := s.userGroupRepo.IsUserInGroup(user.ID, req.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user membership: %w", err)
	}

	if !isInGroup {
		return &dto.AuthResponse{Authorized: false}, nil
	}

	role, err := s.userGroupRepo.GetUserRole(user.ID, req.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}

	return &dto.AuthResponse{
		Authorized: true,
		User:       *utils.UserToCommon(user),
		Group:      *utils.GroupToCommon(group),
		Role:       role,
	}, nil
}

func (s *authService) CheckPermission(telegramID int64, groupID uint, requiredRole string) error {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("пользователь не найден")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	isInGroup, err := s.userGroupRepo.IsUserInGroup(user.ID, groupID)
	if err != nil {
		return fmt.Errorf("failed to check user membership: %w", err)
	}

	if !isInGroup {
		return fmt.Errorf("пользователь не состоит в группе")
	}

	userRole, err := s.userGroupRepo.GetUserRole(user.ID, groupID)
	if err != nil {
		return fmt.Errorf("failed to get user role: %w", err)
	}

	if !s.hasPermission(userRole, requiredRole) {
		return fmt.Errorf("недостаточно прав доступа")
	}

	return nil
}

func (s *authService) GetUserRole(telegramID int64, groupID uint) (string, error) {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("пользователь не найден")
		}
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	role, err := s.userGroupRepo.GetUserRole(user.ID, groupID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("пользователь не состоит в группе")
		}
		return "", fmt.Errorf("failed to get user role: %w", err)
	}

	return role, nil
}

func (s *authService) IsUserInGroup(telegramID int64, groupID uint) (bool, error) {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	return s.userGroupRepo.IsUserInGroup(user.ID, groupID)
}

func (s *authService) hasPermission(userRole, requiredRole string) bool {
	roleHierarchy := map[string]int{
		models.RoleMember:    1,
		models.RoleModerator: 2,
		models.RoleAdmin:     3,
	}

	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}

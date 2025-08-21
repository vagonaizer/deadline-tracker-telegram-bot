package services

import (
	"deadline-bot/internal/dto"
	"deadline-bot/internal/repositories"
	"deadline-bot/pkg/utils"
	"fmt"

	"gorm.io/gorm"
)

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) CreateUser(req *dto.UserRequest) (*dto.UserResponse, error) {
	if err := utils.ValidateUsername(req.Username); err != nil {
		return nil, err
	}

	existingUser, err := s.userRepo.GetByTelegramID(req.TelegramID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		return utils.UserToResponse(existingUser), nil
	}

	user := utils.UserRequestToModel(req)

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return utils.UserToResponse(user), nil
}

func (s *userService) GetUserByTelegramID(telegramID int64) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return utils.UserToResponse(user), nil
}

func (s *userService) GetUserByID(id uint) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return utils.UserToResponse(user), nil
}

func (s *userService) UpdateUser(id uint, req *dto.UserRequest) (*dto.UserResponse, error) {
	if err := utils.ValidateUsername(req.Username); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Username = req.Username
	user.FirstName = req.FirstName
	user.LastName = req.LastName

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return utils.UserToResponse(user), nil
}

func (s *userService) DeleteUser(id uint) error {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("пользователь не найден")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.userRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

package services

import (
	"deadline-bot/internal/config"
	"deadline-bot/internal/dto"
	"deadline-bot/internal/models"
	"deadline-bot/internal/repositories"
	"deadline-bot/pkg/utils"

	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

type fileService struct {
	fileRepo      repositories.FileRepository
	taskRepo      repositories.TaskRepository
	userRepo      repositories.UserRepository
	userGroupRepo repositories.UserGroupRepository
	authService   AuthService
	config        *config.FilesConfig
}

func NewFileService(
	fileRepo repositories.FileRepository,
	taskRepo repositories.TaskRepository,
	userRepo repositories.UserRepository,
	userGroupRepo repositories.UserGroupRepository,
	authService AuthService,
	config *config.FilesConfig,
) FileService {
	return &fileService{
		fileRepo:      fileRepo,
		taskRepo:      taskRepo,
		userRepo:      userRepo,
		userGroupRepo: userGroupRepo,
		authService:   authService,
		config:        config,
	}
}

func (s *fileService) UploadFile(telegramID int64, req *dto.FileUploadRequest) (*dto.FileResponse, error) {
	task, err := s.taskRepo.GetByID(req.TaskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("задача не найдена")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if err := s.authService.CheckPermission(telegramID, task.GroupID, models.RoleMember); err != nil {
		return nil, err
	}

	if err := s.ValidateFile(req.FileName, int64(len(req.FileData)), req.MimeType); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(s.config.UploadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	fileName := s.generateFileName(req.FileName)
	filePath := filepath.Join(s.config.UploadPath, fileName)

	if err := ioutil.WriteFile(filePath, req.FileData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	file := &models.File{
		TaskID:   req.TaskID,
		FileName: req.FileName,
		FilePath: filePath,
		FileSize: int64(len(req.FileData)),
		MimeType: req.MimeType,
	}

	if err := s.fileRepo.Create(file); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save file info: %w", err)
	}

	return utils.FileToResponse(file), nil
}

func (s *fileService) GetTaskFiles(taskID uint) ([]dto.FileResponse, error) {
	files, err := s.fileRepo.GetByTaskID(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task files: %w", err)
	}

	responses := make([]dto.FileResponse, len(files))
	for i, file := range files {
		responses[i] = *utils.FileToResponse(&file)
	}

	return responses, nil
}

func (s *fileService) GetFileByID(id uint) (*dto.FileResponse, error) {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("файл не найден")
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return utils.FileToResponse(file), nil
}

func (s *fileService) DeleteFile(telegramID int64, fileID uint) error {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("файл не найден")
		}
		return fmt.Errorf("failed to get file: %w", err)
	}

	task, err := s.taskRepo.GetByID(file.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	canDelete, err := s.canUserDeleteFile(user.ID, task)
	if err != nil {
		return err
	}

	if !canDelete {
		return fmt.Errorf("недостаточно прав для удаления файла")
	}

	if err := s.fileRepo.Delete(fileID); err != nil {
		return fmt.Errorf("failed to delete file from database: %w", err)
	}

	if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file from disk: %w", err)
	}

	return nil
}

func (s *fileService) GetFilePath(id uint) (string, error) {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("файл не найден")
		}
		return "", fmt.Errorf("failed to get file: %w", err)
	}

	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("файл не найден на диске")
	}

	return file.FilePath, nil
}

func (s *fileService) ValidateFile(fileName string, fileSize int64, mimeType string) error {
	return utils.ValidateFile(fileName, fileSize, mimeType, s.config.AllowedExts, s.config.MaxSizeMB)
}

func (s *fileService) generateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d_%s%s", timestamp, "file", ext)
}

func (s *fileService) canUserDeleteFile(userID uint, task *models.Task) (bool, error) {
	if task.CreatedBy == userID {
		return true, nil
	}

	role, err := s.userGroupRepo.GetUserRole(userID, task.GroupID)
	if err != nil {
		return false, fmt.Errorf("failed to get user role: %w", err)
	}

	return role == models.RoleAdmin || role == models.RoleModerator, nil
}

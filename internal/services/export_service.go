package services

import (
	"deadline-bot/internal/dto"
	"deadline-bot/internal/models"
	"deadline-bot/internal/repositories"
	"deadline-bot/pkg/utils"
	"fmt"
)

type exportService struct {
	taskRepo      repositories.TaskRepository
	userRepo      repositories.UserRepository
	userGroupRepo repositories.UserGroupRepository
	authService   AuthService
}

func NewExportService(
	taskRepo repositories.TaskRepository,
	userRepo repositories.UserRepository,
	userGroupRepo repositories.UserGroupRepository,
	authService AuthService,
) ExportService {
	return &exportService{
		taskRepo:      taskRepo,
		userRepo:      userRepo,
		userGroupRepo: userGroupRepo,
		authService:   authService,
	}
}

func (s *exportService) ExportToExcel(telegramID int64, req *dto.ExportRequest) (*dto.ExportResponse, error) {
	if err := s.authService.CheckPermission(telegramID, req.GroupID, models.RoleMember); err != nil {
		return nil, err
	}

	data, err := s.PrepareExportData(req.GroupID, &dto.TaskFilterRequest{
		Status:   req.Status,
		FromDate: req.FromDate,
		ToDate:   req.ToDate,
	})
	if err != nil {
		return nil, err
	}

	csvData, err := utils.GenerateCSV(data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CSV: %w", err)
	}

	return &dto.ExportResponse{
		FileName: "tasks_export.csv",
		FileData: csvData,
		MimeType: "text/csv",
	}, nil
}

func (s *exportService) ExportToPDF(telegramID int64, req *dto.ExportRequest) (*dto.ExportResponse, error) {
	if err := s.authService.CheckPermission(telegramID, req.GroupID, models.RoleMember); err != nil {
		return nil, err
	}

	data, err := s.PrepareExportData(req.GroupID, &dto.TaskFilterRequest{
		Status:   req.Status,
		FromDate: req.FromDate,
		ToDate:   req.ToDate,
	})
	if err != nil {
		return nil, err
	}

	// Получаем название группы
	group, err := s.getGroupByID(req.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	pdfData, err := utils.GeneratePDF(data, group.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return &dto.ExportResponse{
		FileName: "tasks_export.pdf",
		FileData: pdfData,
		MimeType: "application/pdf",
	}, nil
}

func (s *exportService) PrepareExportData(groupID uint, filter *dto.TaskFilterRequest) ([]dto.ExportTaskData, error) {
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

	exportData := make([]dto.ExportTaskData, len(tasks))
	for i, task := range tasks {
		exportData[i] = *utils.TaskToExportData(&task)
	}

	return exportData, nil
}

func (s *exportService) getGroupByID(groupID uint) (*models.Group, error) {
	// Это временная заглушка, в реальности нужен GroupRepository
	return &models.Group{
		ID:   groupID,
		Name: "Группа",
	}, nil
}

func (s *exportService) getFilteredTasks(groupID uint, filter *dto.TaskFilterRequest) ([]models.Task, error) {
	if filter.Status != nil {
		return s.taskRepo.GetByStatus(groupID, *filter.Status)
	}

	if filter.FromDate != nil && filter.ToDate != nil {
		return s.taskRepo.GetByDeadlineRange(groupID, *filter.FromDate, *filter.ToDate)
	}

	return s.taskRepo.GetByGroupID(groupID)
}

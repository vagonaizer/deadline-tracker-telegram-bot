package utils

import (
	"deadline-bot/internal/dto"
	"deadline-bot/internal/models"
)

func UserToResponse(user *models.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:         user.ID,
		TelegramID: user.TelegramID,
		Username:   user.Username,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		CreatedAt:  user.CreatedAt,
	}
}

func UserToCommon(user *models.User) *dto.UserCommon {
	return &dto.UserCommon{
		ID:         user.ID,
		TelegramID: user.TelegramID,
		Username:   user.Username,
		FirstName:  user.FirstName,
	}
}

func GroupToResponse(group *models.Group) *dto.GroupResponse {
	return &dto.GroupResponse{
		ID:        group.ID,
		Name:      group.Name,
		Login:     group.Login,
		Creator:   *UserToCommon(&group.Creator),
		CreatedAt: group.CreatedAt,
	}
}

func GroupToCommon(group *models.Group) *dto.GroupCommon {
	return &dto.GroupCommon{
		ID:    group.ID,
		Name:  group.Name,
		Login: group.Login,
	}
}

func TaskToResponse(task *models.Task) *dto.TaskResponse {
	response := &dto.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Deadline:    task.Deadline,
		Priority:    task.Priority,
		Status:      task.Status,
		Group:       *GroupToCommon(&task.Group),
		Creator:     *UserToCommon(&task.Creator),
		CreatedAt:   task.CreatedAt,
	}

	if task.Assignee != nil {
		assignee := UserToCommon(task.Assignee)
		response.Assignee = assignee
	}

	if len(task.Files) > 0 {
		files := make([]dto.FileCommon, len(task.Files))
		for i, file := range task.Files {
			files[i] = *FileToCommon(&file)
		}
		response.Files = files
	}

	return response
}

func TaskToCommon(task *models.Task) *dto.TaskCommon {
	return &dto.TaskCommon{
		ID:       task.ID,
		Title:    task.Title,
		Deadline: task.Deadline,
		Priority: task.Priority,
		Status:   task.Status,
	}
}

func FileToResponse(file *models.File) *dto.FileResponse {
	return &dto.FileResponse{
		ID:        file.ID,
		TaskID:    file.TaskID,
		FileName:  file.FileName,
		FileSize:  file.FileSize,
		MimeType:  file.MimeType,
		CreatedAt: file.CreatedAt,
	}
}

func FileToCommon(file *models.File) *dto.FileCommon {
	return &dto.FileCommon{
		ID:       file.ID,
		FileName: file.FileName,
		FileSize: file.FileSize,
		MimeType: file.MimeType,
	}
}

func NotificationToResponse(notification *models.Notification) *dto.NotificationResponse {
	return &dto.NotificationResponse{
		ID:        notification.ID,
		User:      *UserToCommon(&notification.User),
		Task:      *TaskToCommon(&notification.Task),
		Type:      notification.Type,
		SentAt:    notification.SentAt,
		CreatedAt: notification.CreatedAt,
	}
}

func UserGroupToMemberResponse(userGroup *models.UserGroup) *dto.GroupMemberResponse {
	return &dto.GroupMemberResponse{
		User:     *UserToCommon(&userGroup.User),
		Role:     userGroup.Role,
		JoinedAt: userGroup.JoinedAt,
	}
}

func TaskToExportData(task *models.Task) *dto.ExportTaskData {
	assignee := ""
	if task.Assignee != nil {
		assignee = task.Assignee.FirstName
		if task.Assignee.LastName != "" {
			assignee += " " + task.Assignee.LastName
		}
		if assignee == "" {
			assignee = task.Assignee.Username
		}
	}

	creator := task.Creator.FirstName
	if task.Creator.LastName != "" {
		creator += " " + task.Creator.LastName
	}
	if creator == "" {
		creator = task.Creator.Username
	}

	return &dto.ExportTaskData{
		Title:       task.Title,
		Description: task.Description,
		Deadline:    task.Deadline,
		Priority:    task.Priority,
		Status:      task.Status,
		Creator:     creator,
		Assignee:    assignee,
		CreatedAt:   task.CreatedAt,
	}
}

func UserRequestToModel(req *dto.UserRequest) *models.User {
	return &models.User{
		TelegramID: req.TelegramID,
		Username:   req.Username,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
	}
}

func CreateGroupRequestToModel(req *dto.CreateGroupRequest, createdBy uint, hashedPassword string) *models.Group {
	return &models.Group{
		Name:      req.Name,
		Login:     req.Login,
		Password:  hashedPassword,
		CreatedBy: createdBy,
	}
}

func CreateTaskRequestToModel(req *dto.CreateTaskRequest, groupID, createdBy uint) *models.Task {
	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Deadline:    req.Deadline,
		Priority:    GetPriorityByDeadline(req.Deadline),
		Status:      models.StatusPending,
		GroupID:     groupID,
		CreatedBy:   createdBy,
	}

	if req.AssignedTo != nil {
		task.AssignedTo = req.AssignedTo
	}

	return task
}

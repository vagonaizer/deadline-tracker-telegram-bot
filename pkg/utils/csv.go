package utils

import (
	"deadline-bot/internal/dto"
	"encoding/csv"
	"fmt"
	"strings"
)

func GenerateCSV(tasks []dto.ExportTaskData) ([]byte, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Заголовки
	headers := []string{
		"Название",
		"Описание",
		"Дедлайн",
		"Приоритет",
		"Статус",
		"Создатель",
		"Назначен",
		"Дата создания",
	}

	if err := writer.Write(headers); err != nil {
		return nil, fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Данные
	for _, task := range tasks {
		record := []string{
			task.Title,
			task.Description,
			task.Deadline.Format("02.01.2006 15:04"),
			getPriorityRussian(task.Priority),
			getStatusRussian(task.Status),
			task.Creator,
			task.Assignee,
			task.CreatedAt.Format("02.01.2006 15:04"),
		}

		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return []byte(buf.String()), nil
}

func getPriorityRussian(priority string) string {
	switch priority {
	case "low":
		return "Низкий"
	case "normal":
		return "Обычный"
	case "high":
		return "Высокий"
	case "critical":
		return "Критический"
	default:
		return priority
	}
}

func getStatusRussian(status string) string {
	switch status {
	case "pending":
		return "В работе"
	case "completed":
		return "Выполнено"
	case "overdue":
		return "Просрочено"
	default:
		return status
	}
}

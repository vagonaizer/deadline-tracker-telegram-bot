package utils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func RespondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{Error: message})
}

func RespondWithErrorDetails(c *gin.Context, code int, message, details string) {
	c.JSON(code, ErrorResponse{
		Error:   message,
		Details: details,
	})
}

func RespondWithSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func RespondWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, SuccessResponse{Message: message})
}

func RespondWithCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

func HandleRepositoryError(err error) (int, string) {
	errMsg := err.Error()

	if strings.Contains(errMsg, "record not found") {
		return http.StatusNotFound, "Запись не найдена"
	}

	if strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique") {
		return http.StatusConflict, "Запись уже существует"
	}

	if strings.Contains(errMsg, "foreign key") {
		return http.StatusBadRequest, "Связанная запись не найдена"
	}

	return http.StatusInternalServerError, "Внутренняя ошибка сервера"
}

func HandleValidationError(err error) (int, string) {
	return http.StatusBadRequest, fmt.Sprintf("Ошибка валидации: %s", err.Error())
}

func HandleAuthorizationError(message string) (int, string) {
	return http.StatusForbidden, message
}

func HandleNotFoundError(resource string) (int, string) {
	return http.StatusNotFound, fmt.Sprintf("%s не найден", resource)
}

func FormatTelegramMessage(title, text string) string {
	var builder strings.Builder

	builder.WriteString("*")
	builder.WriteString(EscapeMarkdown(title))
	builder.WriteString("*\n\n")
	builder.WriteString(EscapeMarkdown(text))

	return builder.String()
}

func FormatTaskMessage(title, description, deadline, priority string) string {
	var builder strings.Builder

	builder.WriteString("📋 *")
	builder.WriteString(EscapeMarkdown(title))
	builder.WriteString("*\n\n")

	if description != "" {
		builder.WriteString("📝 ")
		builder.WriteString(EscapeMarkdown(description))
		builder.WriteString("\n\n")
	}

	builder.WriteString("🗓 *Дедлайн:* ")
	builder.WriteString(EscapeMarkdown(deadline))
	builder.WriteString("\n")

	priorityEmoji := getPriorityEmoji(priority)
	builder.WriteString(priorityEmoji)
	builder.WriteString(" *Приоритет:* ")
	builder.WriteString(EscapeMarkdown(priority))

	return builder.String()
}

func EscapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"*", "\\*",
		"_", "\\_",
		"`", "\\`",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

func getPriorityEmoji(priority string) string {
	switch priority {
	case "low":
		return "🟢"
	case "normal":
		return "🟡"
	case "high":
		return "🟠"
	case "critical":
		return "🔴"
	default:
		return "⚪"
	}
}

func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

package utils

import (
	"fmt"
	"mime"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var (
	loginRegex    = regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{1,100}$`)
)

func ValidateGroupLogin(login string) error {
	if len(login) < 3 || len(login) > 50 {
		return fmt.Errorf("login must be between 3 and 50 characters")
	}

	if !loginRegex.MatchString(login) {
		return fmt.Errorf("login can only contain letters, numbers and underscores")
	}

	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	if len(password) > 100 {
		return fmt.Errorf("password must be no more than 100 characters")
	}

	hasLetter := false
	hasDigit := false

	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
	}

	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}

	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	return nil
}

func ValidateUsername(username string) error {
	if len(username) == 0 {
		return nil
	}

	if len(username) > 100 {
		return fmt.Errorf("username must be no more than 100 characters")
	}

	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers and underscores")
	}

	return nil
}

func ValidateTaskTitle(title string) error {
	title = strings.TrimSpace(title)
	if len(title) == 0 {
		return fmt.Errorf("task title cannot be empty")
	}

	if len(title) > 200 {
		return fmt.Errorf("task title must be no more than 200 characters")
	}

	return nil
}

func ValidateTaskDescription(description string) error {
	if len(description) > 1000 {
		return fmt.Errorf("task description must be no more than 1000 characters")
	}

	return nil
}

func ValidateDeadline(deadline time.Time) error {
	if deadline.Before(time.Now()) {
		return fmt.Errorf("deadline cannot be in the past")
	}

	maxFuture := time.Now().AddDate(5, 0, 0)
	if deadline.After(maxFuture) {
		return fmt.Errorf("deadline cannot be more than 5 years in the future")
	}

	return nil
}

func ValidateFile(fileName string, fileSize int64, mimeType string, allowedExts []string, maxSizeMB int) error {
	if len(fileName) == 0 {
		return fmt.Errorf("file name cannot be empty")
	}

	if len(fileName) > 255 {
		return fmt.Errorf("file name must be no more than 255 characters")
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		return fmt.Errorf("file must have an extension")
	}

	allowed := false
	for _, allowedExt := range allowedExts {
		if ext == strings.ToLower(allowedExt) {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("file type %s is not allowed", ext)
	}

	maxSize := int64(maxSizeMB) * 1024 * 1024
	if fileSize > maxSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size %d MB", fileSize, maxSizeMB)
	}

	if fileSize <= 0 {
		return fmt.Errorf("file size must be greater than 0")
	}

	expectedMimeType := mime.TypeByExtension(ext)
	if expectedMimeType != "" && mimeType != expectedMimeType {
		return fmt.Errorf("mime type %s does not match file extension %s", mimeType, ext)
	}

	return nil
}

func ValidateRole(role string) error {
	validRoles := []string{"admin", "moderator", "member"}

	for _, validRole := range validRoles {
		if role == validRole {
			return nil
		}
	}

	return fmt.Errorf("invalid role: %s", role)
}

func ValidateStatus(status string) error {
	validStatuses := []string{"pending", "completed", "overdue"}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}

	return fmt.Errorf("invalid status: %s", status)
}

func ValidatePriority(priority string) error {
	validPriorities := []string{"low", "normal", "high", "critical"}

	for _, validPriority := range validPriorities {
		if priority == validPriority {
			return nil
		}
	}

	return fmt.Errorf("invalid priority: %s", priority)
}

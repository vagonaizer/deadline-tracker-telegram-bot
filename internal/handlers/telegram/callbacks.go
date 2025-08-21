package telegram

import (
	"deadline-bot/internal/dto"
	"deadline-bot/pkg/utils"

	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleCallback(callbackQuery *tgbotapi.CallbackQuery) {
	data := callbackQuery.Data
	userID := callbackQuery.From.ID
	chatID := callbackQuery.Message.Chat.ID
	messageID := callbackQuery.Message.MessageID

	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	b.api.Request(callback)

	parts := strings.Split(data, "_")
	if len(parts) < 2 {
		return
	}

	action := parts[0]

	switch action {
	case "select":
		if len(parts) >= 3 && parts[1] == "group" {
			groupID, err := strconv.ParseUint(parts[2], 10, 32)
			if err != nil {
				b.sendMessage(chatID, "Ошибка при выборе группы")
				return
			}
			b.handleSelectGroup(chatID, messageID, userID, uint(groupID))
		}
	case "export":
		if len(parts) >= 3 {
			format := parts[1]
			groupID, err := strconv.ParseUint(parts[2], 10, 32)
			if err != nil {
				b.sendMessage(chatID, "Ошибка при экспорте")
				return
			}
			b.handleExport(chatID, userID, uint(groupID), format)
		}
	case "task":
		if len(parts) >= 3 {
			subAction := parts[1]
			taskID, err := strconv.ParseUint(parts[2], 10, 32)
			if err != nil {
				b.sendMessage(chatID, "Ошибка при обработке задачи")
				return
			}
			b.handleTaskAction(chatID, userID, uint(taskID), subAction)
		}
	}
}

func (b *Bot) handleSelectGroup(chatID int64, messageID int, userID int64, groupID uint) {
	group, err := b.groupService.GetGroupByID(groupID)
	if err != nil {
		b.editMessage(chatID, messageID, "❌ Ошибка при выборе группы: "+err.Error(), nil)
		return
	}

	isInGroup, err := b.authService.IsUserInGroup(userID, groupID)
	if err != nil || !isInGroup {
		b.editMessage(chatID, messageID, "❌ Вы не состоите в этой группе", nil)
		return
	}

	session := getUserSession(userID)
	if session == nil {
		session = &UserSession{Data: make(map[string]interface{})}
	}
	session.GroupID = groupID
	setUserSession(userID, session)

	role, err := b.authService.GetUserRole(userID, groupID)
	if err != nil {
		role = "неизвестно"
	}

	roleEmoji := b.getRoleEmoji(role)

	text := fmt.Sprintf(`✅ *Группа выбрана*

📋 *Название:* %s
👤 *Логин:* @%s  
%s *Ваша роль:* %s

Теперь вы можете:
/new_task - создать задачу
/tasks - посмотреть задачи
/export - экспорт задач`,
		utils.EscapeMarkdown(group.Name),
		group.Login,
		roleEmoji,
		role)

	b.editMessage(chatID, messageID, text, nil)
}

func (b *Bot) handleExport(chatID int64, userID int64, groupID uint, format string) {
	req := &dto.ExportRequest{
		GroupID: groupID,
		Format:  format,
	}

	var result *dto.ExportResponse
	var err error

	switch format {
	case "excel":
		result, err = b.exportService.ExportToExcel(userID, req)
	case "pdf":
		result, err = b.exportService.ExportToPDF(userID, req)
	default:
		b.sendMessage(chatID, "❌ Неподдерживаемый формат экспорта")
		return
	}

	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка при экспорте: "+err.Error())
		return
	}

	// Отправляем файл как документ
	if len(result.FileData) > 0 {
		document := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{
			Name:  result.FileName,
			Bytes: result.FileData,
		})
		document.Caption = fmt.Sprintf("📋 Экспорт задач готов!\n📄 Файл: %s", result.FileName)

		if _, err := b.api.Send(document); err != nil {
			b.sendMessage(chatID, "❌ Ошибка при отправке файла: "+err.Error())
			return
		}
	} else {
		text := fmt.Sprintf("📋 Экспорт задач готов!\n📄 Файл: %s", result.FileName)
		msg := tgbotapi.NewMessage(chatID, text)
		b.api.Send(msg)
	}
}

func (b *Bot) handleTaskAction(chatID int64, userID int64, taskID uint, action string) {
	switch action {
	case "complete":
		err := b.taskService.CompleteTask(userID, taskID)
		if err != nil {
			b.sendMessage(chatID, "❌ Ошибка при завершении задачи: "+err.Error())
			return
		}
		b.sendMessage(chatID, "✅ Задача отмечена как выполненная!")

	case "delete":
		err := b.taskService.DeleteTask(userID, taskID)
		if err != nil {
			b.sendMessage(chatID, "❌ Ошибка при удалении задачи: "+err.Error())
			return
		}
		b.sendMessage(chatID, "🗑 Задача удалена")

	case "details":
		task, err := b.taskService.GetTaskByID(taskID)
		if err != nil {
			b.sendMessage(chatID, "❌ Ошибка при получении задачи: "+err.Error())
			return
		}
		b.showTaskDetails(chatID, task)
	}
}

func (b *Bot) showTaskDetails(chatID int64, task *dto.TaskResponse) {
	priorityEmoji := b.getPriorityEmoji(task.Priority)
	statusEmoji := b.getStatusEmoji(task.Status)

	var text strings.Builder
	text.WriteString(fmt.Sprintf("%s %s *%s*\n\n", priorityEmoji, statusEmoji, utils.EscapeMarkdown(task.Title)))

	if task.Description != "" {
		text.WriteString(fmt.Sprintf("📝 *Описание:*\n%s\n\n", utils.EscapeMarkdown(task.Description)))
	}

	text.WriteString(fmt.Sprintf("🗓 *Дедлайн:* %s\n", utils.FormatDeadline(task.Deadline)))
	text.WriteString(fmt.Sprintf("🎯 *Приоритет:* %s\n", task.Priority))
	text.WriteString(fmt.Sprintf("📊 *Статус:* %s\n", task.Status))
	text.WriteString(fmt.Sprintf("👤 *Создатель:* %s\n", task.Creator.FirstName))

	if task.Assignee != nil {
		text.WriteString(fmt.Sprintf("🎯 *Назначено:* %s\n", task.Assignee.FirstName))
	}

	text.WriteString(fmt.Sprintf("📅 *Создано:* %s\n", utils.FormatDeadline(task.CreatedAt)))

	if len(task.Files) > 0 {
		text.WriteString(fmt.Sprintf("\n📎 *Файлы (%d):*\n", len(task.Files)))
		for _, file := range task.Files {
			text.WriteString(fmt.Sprintf("• %s (%s)\n",
				utils.EscapeMarkdown(file.FileName),
				utils.FormatFileSize(file.FileSize)))
		}
	}

	timeLeft := utils.FormatDuration(task.Deadline.Sub(time.Now()))
	text.WriteString(fmt.Sprintf("\n⏰ *Осталось:* %s", timeLeft))

	keyboard := tgbotapi.NewInlineKeyboardMarkup()

	if task.Status != "completed" {
		completeBtn := tgbotapi.NewInlineKeyboardButtonData("✅ Завершить", fmt.Sprintf("task_complete_%d", task.ID))
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(completeBtn))
	}

	deleteBtn := tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить", fmt.Sprintf("task_delete_%d", task.ID))
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(deleteBtn))

	b.sendMessageWithKeyboard(chatID, text.String(), keyboard)
}

func (b *Bot) getRoleEmoji(role string) string {
	switch role {
	case "admin":
		return "👑"
	case "moderator":
		return "🛡"
	case "member":
		return "👤"
	default:
		return "❓"
	}
}

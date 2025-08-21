package telegram

import (
	"deadline-bot/internal/dto"
	"deadline-bot/pkg/utils"

	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleGroupNameInput(message *tgbotapi.Message, session *UserSession) {
	name := strings.TrimSpace(message.Text)
	if len(name) == 0 || len(name) > 200 {
		b.sendMessage(message.Chat.ID, "Название группы должно содержать от 1 до 200 символов. Попробуйте еще раз:")
		return
	}

	session.Data["name"] = name
	session.State = "creating_group_login"

	b.sendMessage(message.Chat.ID, "Введите логин группы (3-50 символов, только буквы, цифры и подчеркивания):")
}

func (b *Bot) handleGroupLoginInput(message *tgbotapi.Message, session *UserSession) {
	login := strings.TrimSpace(message.Text)

	if err := utils.ValidateGroupLogin(login); err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка: "+err.Error()+"\nПопробуйте еще раз:")
		return
	}

	session.Data["login"] = login
	session.State = "creating_group_password"

	b.sendMessage(message.Chat.ID, "Введите пароль группы (минимум 6 символов, должен содержать букву и цифру):")
}

func (b *Bot) handleGroupPasswordInput(message *tgbotapi.Message, session *UserSession) {
	password := message.Text

	if err := utils.ValidatePassword(password); err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка: "+err.Error()+"\nПопробуйте еще раз:")
		return
	}

	req := &dto.CreateGroupRequest{
		Name:     session.Data["name"].(string),
		Login:    session.Data["login"].(string),
		Password: password,
	}

	group, err := b.groupService.CreateGroup(message.From.ID, req)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при создании группы: "+err.Error())
		clearUserSession(message.From.ID)
		return
	}

	session.GroupID = group.ID
	clearUserSession(message.From.ID)
	setUserSession(message.From.ID, &UserSession{GroupID: group.ID, Data: make(map[string]interface{})})

	text := "✅ Группа успешно создана!\n\n📋 Название: " + group.Name + "\n👤 Логин: @" + group.Login + "\n\nВы автоматически назначены администратором группы.\nТеперь можете создавать задачи командой /new_task"

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.api.Send(msg)
}

func (b *Bot) handleConnectLoginInput(message *tgbotapi.Message, session *UserSession) {
	login := strings.TrimSpace(message.Text)

	if len(login) == 0 {
		b.sendMessage(message.Chat.ID, "Логин не может быть пустым. Попробуйте еще раз:")
		return
	}

	session.Data["login"] = login
	session.State = "connecting_password"

	b.sendMessage(message.Chat.ID, "Введите пароль группы:")
}

func (b *Bot) handleConnectPasswordInput(message *tgbotapi.Message, session *UserSession) {
	password := message.Text

	req := &dto.ConnectGroupRequest{
		Login:    session.Data["login"].(string),
		Password: password,
	}

	group, err := b.groupService.ConnectToGroup(message.From.ID, req)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при подключении: "+err.Error())
		clearUserSession(message.From.ID)
		return
	}

	session.GroupID = group.ID
	clearUserSession(message.From.ID)
	setUserSession(message.From.ID, &UserSession{GroupID: group.ID, Data: make(map[string]interface{})})

	text := "✅ Успешно подключились к группе!\n\n📋 Название: " + group.Name + "\n👤 Логин: @" + group.Login + "\n\nТеперь можете просматривать задачи командой /tasks"

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.api.Send(msg)
}

func (b *Bot) handleTaskTitleInput(message *tgbotapi.Message, session *UserSession) {
	title := strings.TrimSpace(message.Text)

	if err := utils.ValidateTaskTitle(title); err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка: "+err.Error()+"\nПопробуйте еще раз:")
		return
	}

	session.Data["title"] = title
	session.State = "creating_task_description"

	b.sendMessage(message.Chat.ID, "Введите описание задачи (или отправьте \"-\" чтобы пропустить):")
}

func (b *Bot) handleTaskDescriptionInput(message *tgbotapi.Message, session *UserSession) {
	description := strings.TrimSpace(message.Text)

	if description == "-" {
		description = ""
	} else if err := utils.ValidateTaskDescription(description); err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка: "+err.Error()+"\nПопробуйте еще раз:")
		return
	}

	session.Data["description"] = description
	session.State = "creating_task_deadline"

	text := `Введите дедлайн задачи в одном из форматов:
• 25.12.2024 23:59
• 25.12.2024
• 2024-12-25 23:59
• 2024-12-25`

	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) handleTaskDeadlineInput(message *tgbotapi.Message, session *UserSession) {
	deadlineStr := strings.TrimSpace(message.Text)

	deadline, err := utils.ParseDeadline(deadlineStr)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при разборе даты: "+err.Error()+"\nПопробуйте еще раз:")
		return
	}

	if err := utils.ValidateDeadline(deadline); err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка: "+err.Error()+"\nПопробуйте еще раз:")
		return
	}

	groupID := session.Data["group_id"].(uint)

	req := &dto.CreateTaskRequest{
		Title:       session.Data["title"].(string),
		Description: session.Data["description"].(string),
		Deadline:    deadline,
	}

	task, err := b.taskService.CreateTask(message.From.ID, groupID, req)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при создании задачи: "+err.Error())
		clearUserSession(message.From.ID)
		return
	}

	clearUserSession(message.From.ID)
	setUserSession(message.From.ID, &UserSession{GroupID: groupID, Data: make(map[string]interface{})})

	priorityEmoji := b.getPriorityEmoji(task.Priority)

	text := `✅ *Задача успешно создана!*

` + priorityEmoji + ` *` + utils.EscapeMarkdown(task.Title) + `*

🗓 *Дедлайн:* ` + utils.FormatDeadline(task.Deadline) + `
🎯 *Приоритет:* ` + task.Priority

	if task.Description != "" {
		text += `
📝 *Описание:* ` + utils.EscapeMarkdown(task.Description)
	}

	b.sendMessage(message.Chat.ID, text)
}

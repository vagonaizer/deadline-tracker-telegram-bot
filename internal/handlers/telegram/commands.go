package telegram

import (
	"deadline-bot/pkg/utils"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleStartCommand(message *tgbotapi.Message) {
	text := `Добро пожаловать в DeadlineBot!

Я помогу вам управлять дедлайнами в группе/классе.

📋 Основные команды:

/help - справка по всем командам
/new_group - создать новую группу
/connect - подключиться к группе
/my_groups - мои группы

Начните с создания группы или подключения к существующей!`

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.api.Send(msg)

	b.sendWelcomeImage(message.Chat.ID)
}

func (b *Bot) handleHelpCommand(message *tgbotapi.Message) {
	text := `📋 Справка по командам DeadlineBot

🏢 Управление группами:

/new_group - создать новую группу
/connect - подключиться к группе по логину и паролю
/my_groups - список ваших групп
/select_group - выбрать активную группу

📋 Управление задачами:

/new_task - создать новую задачу
/tasks - показать задачи группы
/export - экспорт задач в Excel/PDF

⚙️ Общие команды:

/cancel - отменить текущую операцию
/help - эта справка

🎯 Приоритеты задач:

🟢 Низкий - более недели до дедлайна
🟡 Обычный - 3-7 дней до дедлайна  
🟠 Высокий - 1-3 дня до дедлайна
🔴 Критический - менее дня или просрочено`

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.api.Send(msg)

	b.sendPrioritiesImage(message.Chat.ID)
}

func (b *Bot) handleNewGroupCommand(message *tgbotapi.Message) {
	updateUserSessionState(message.From.ID, "creating_group_name")

	text := `🆕 Создание новой группы

Введите название группы:`

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.api.Send(msg)

	b.sendCreateGroupImage(message.Chat.ID)
}

func (b *Bot) handleConnectCommand(message *tgbotapi.Message) {
	updateUserSessionState(message.From.ID, "connecting_login")

	text := `🔗 Подключение к группе

Введите логин группы:`

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.api.Send(msg)

	b.sendConnectGroupImage(message.Chat.ID)
}

func (b *Bot) handleMyGroupsCommand(message *tgbotapi.Message) {
	groups, err := b.groupService.GetUserGroups(message.From.ID)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при получении групп: "+err.Error())
		return
	}

	if len(groups) == 0 {
		text := "📝 Вы не состоите ни в одной группе\n\nИспользуйте:\n/new_group - чтобы создать новую группу\n/connect - чтобы подключиться к существующей"

		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		b.api.Send(msg)
		return
	}

	var text strings.Builder
	text.WriteString("📋 Ваши группы:\n\n")

	for i, group := range groups {
		text.WriteString(fmt.Sprintf("%d. %s (@%s)\n", i+1, group.Name, group.Login))
	}

	text.WriteString("\nИспользуйте /select_group для выбора активной группы")

	msg := tgbotapi.NewMessage(message.Chat.ID, text.String())
	b.api.Send(msg)
}

func (b *Bot) handleSelectGroupCommand(message *tgbotapi.Message) {
	groups, err := b.groupService.GetUserGroups(message.From.ID)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при получении групп: "+err.Error())
		return
	}

	if len(groups) == 0 {
		b.sendMessage(message.Chat.ID, "Вы не состоите ни в одной группе. Используйте /my_groups")
		return
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup()
	for _, group := range groups {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s (@%s)", group.Name, group.Login),
			fmt.Sprintf("select_group_%d", group.ID),
		)
		row := tgbotapi.NewInlineKeyboardRow(button)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "🎯 Выберите группу:")
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleNewTaskCommand(message *tgbotapi.Message) {
	session := getUserSession(message.From.ID)
	if session == nil || session.GroupID == 0 {
		text := "❌ Сначала выберите группу командой /select_group\n\nИли создайте новую группу: /new_group"
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		b.api.Send(msg)
		return
	}

	updateUserSessionState(message.From.ID, "creating_task_title")
	session = getUserSession(message.From.ID)
	session.Data["group_id"] = session.GroupID

	text := `📝 Создание новой задачи

Введите название задачи:`

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.api.Send(msg)

	b.sendCreateTaskImage(message.Chat.ID)
}

func (b *Bot) handleTasksCommand(message *tgbotapi.Message) {
	session := getUserSession(message.From.ID)
	if session == nil || session.GroupID == 0 {
		text := "❌ Сначала выберите группу командой /select_group\n\nИли создайте новую группу: /new_group"
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		b.api.Send(msg)
		return
	}

	tasks, err := b.taskService.GetGroupTasks(session.GroupID, nil)
	if err != nil {
		text := "❌ Ошибка при получении задач: " + err.Error()
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		b.api.Send(msg)
		return
	}

	if len(tasks.Tasks) == 0 {
		text := "📝 В группе пока нет задач\n\nИспользуйте /new_task чтобы создать первую задачу"

		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		b.api.Send(msg)
		return
	}

	var text strings.Builder
	text.WriteString("📋 Задачи группы:\n\n")

	for i, task := range tasks.Tasks {
		priorityEmoji := b.getPriorityEmoji(task.Priority)
		statusEmoji := b.getStatusEmoji(task.Status)

		// Название задачи с приоритетом и статусом
		text.WriteString(fmt.Sprintf("Название: %s %s %s\n", priorityEmoji, statusEmoji, task.Title))

		// Дедлайн
		text.WriteString(fmt.Sprintf("Дедлайн: %s\n", utils.FormatDeadline(task.Deadline)))

		// Описание
		if task.Description != "" && len(task.Description) > 0 {
			desc := task.Description
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			text.WriteString(fmt.Sprintf("Описание: %s\n", desc))
		} else {
			text.WriteString("Описание: —\n")
		}

		// Добавляем разделитель между задачами, кроме последней
		if i < len(tasks.Tasks)-1 {
			text.WriteString("\n" + strings.Repeat("─", 30) + "\n\n")
		}
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text.String())
	b.api.Send(msg)

	b.sendTaskManagementImage(message.Chat.ID)
}

func (b *Bot) handleExportCommand(message *tgbotapi.Message) {
	session := getUserSession(message.From.ID)
	if session == nil || session.GroupID == 0 {
		text := "❌ Сначала выберите группу командой /select_group\n\nИли создайте новую группу: /new_group"
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		b.api.Send(msg)
		return
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Excel", fmt.Sprintf("export_excel_%d", session.GroupID)),
			tgbotapi.NewInlineKeyboardButtonData("📄 PDF", fmt.Sprintf("export_pdf_%d", session.GroupID)),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, "📋 Выберите формат экспорта:")
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)

	b.sendExportDemoImage(message.Chat.ID)
}

func (b *Bot) handleSessionInput(message *tgbotapi.Message, session *UserSession) {
	switch session.State {
	case "creating_group_name":
		b.handleGroupNameInput(message, session)
	case "creating_group_login":
		b.handleGroupLoginInput(message, session)
	case "creating_group_password":
		b.handleGroupPasswordInput(message, session)
	case "connecting_login":
		b.handleConnectLoginInput(message, session)
	case "connecting_password":
		b.handleConnectPasswordInput(message, session)
	case "creating_task_title":
		b.handleTaskTitleInput(message, session)
	case "creating_task_description":
		b.handleTaskDescriptionInput(message, session)
	case "creating_task_deadline":
		b.handleTaskDeadlineInput(message, session)
	default:
		clearUserSession(message.From.ID)
		b.sendMessage(message.Chat.ID, "Неизвестное состояние. Операция отменена.")
	}
}

func (b *Bot) getPriorityEmoji(priority string) string {
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

func (b *Bot) getStatusEmoji(status string) string {
	switch status {
	case "pending":
		return "⏳"
	case "completed":
		return "✅"
	case "overdue":
		return "❌"
	default:
		return "❓"
	}
}

package telegram

import (
	"deadline-bot/internal/config"
	"deadline-bot/internal/dto"
	"deadline-bot/internal/services"
	"deadline-bot/pkg/utils"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api                 *tgbotapi.BotAPI
	userService         services.UserService
	groupService        services.GroupService
	taskService         services.TaskService
	authService         services.AuthService
	fileService         services.FileService
	exportService       services.ExportService
	notificationService services.NotificationService
}

type UserSession struct {
	State     string
	Data      map[string]interface{}
	GroupID   uint
	MessageID int
}

var userSessions = make(map[int64]*UserSession)

func NewBot(
	config *config.BotConfig,
	userService services.UserService,
	groupService services.GroupService,
	taskService services.TaskService,
	authService services.AuthService,
	fileService services.FileService,
	exportService services.ExportService,
	notificationService services.NotificationService,
) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}

	api.Debug = config.Debug

	bot := &Bot{
		api:                 api,
		userService:         userService,
		groupService:        groupService,
		taskService:         taskService,
		authService:         authService,
		fileService:         fileService,
		exportService:       exportService,
		notificationService: notificationService,
	}

	// Устанавливаем себя как отправителя уведомлений
	notificationService.SetNotificationSender(bot)

	if err := bot.setCommands(); err != nil {
		log.Printf("Warning: Failed to set bot commands: %v", err)
	}

	return bot, nil
}

func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		go b.handleUpdate(update)
	}

	return nil
}

func (b *Bot) Stop() {
	b.api.StopReceivingUpdates()
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		b.handleMessage(update.Message)
	} else if update.CallbackQuery != nil {
		b.handleCallbackQuery(update.CallbackQuery)
	}
}

func (b *Bot) handleTestNotificationsCommand(message *tgbotapi.Message) {
	// Показываем варианты тестирования
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 Проверить уведомления", "check_notifications"),
			tgbotapi.NewInlineKeyboardButtonData("🧪 Создать тестовое", "create_test_notification"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Мои уведомления", "my_notifications"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, "🔔 Тестирование уведомлений:")
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID

	if message.IsCommand() {
		b.handleCommand(message)
		return
	}

	if message.Document != nil {
		b.handleDocument(message)
		return
	}

	session := getUserSession(userID)
	if session != nil && session.State != "" {
		b.handleSessionInput(message, session)
		return
	}

	b.sendMessage(message.Chat.ID, "Используйте команды для работы с ботом. /help для справки.")
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	userID := message.From.ID

	b.ensureUserExists(message.From)

	switch message.Command() {
	case "start":
		b.handleStartCommand(message)
	case "help":
		b.handleHelpCommand(message)
	case "new_group", "newgroup":
		b.handleNewGroupCommand(message)
	case "connect":
		b.handleConnectCommand(message)
	case "my_groups", "mygroups":
		b.handleMyGroupsCommand(message)
	case "select_group", "selectgroup":
		b.handleSelectGroupCommand(message)
	case "new_task", "newtask":
		b.handleNewTaskCommand(message)
	case "tasks":
		b.handleTasksCommand(message)
	case "export":
		b.handleExportCommand(message)
	case "test_notifications":
		b.handleTestNotificationsCommand(message)
	case "create_test_task":
		b.handleUploadFileCommand(message)
	case "task_files":
		b.handleTaskFilesCommand(message)
	case "download_file":
		b.handleDownloadFileCommand(message)
	case "cancel":
		clearUserSession(userID)
		b.sendMessage(message.Chat.ID, "Операция отменена.")
	default:
		b.sendMessage(message.Chat.ID, "Неизвестная команда. Используйте /help для справки.")
	}
}

func (b *Bot) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	data := callbackQuery.Data

	// Проверяем callbacks для уведомлений
	if data == "check_notifications" || data == "create_test_notification" || data == "my_notifications" {
		b.handleNotificationCallback(callbackQuery)
		return
	}

	// Остальные callbacks
	b.handleCallback(callbackQuery)
}

func (b *Bot) handleNotificationCallback(callbackQuery *tgbotapi.CallbackQuery) {
	data := callbackQuery.Data

	switch data {
	case "check_notifications":
		b.checkAndSendNotifications(callbackQuery)
	case "create_test_notification":
		b.createTestNotification(callbackQuery)
	case "my_notifications":
		b.showMyNotifications(callbackQuery)
	}
}

func (b *Bot) checkAndSendNotifications(callbackQuery *tgbotapi.CallbackQuery) {
	// Отвечаем на callback
	callback := tgbotapi.NewCallback(callbackQuery.ID, "Проверяю уведомления...")
	b.api.Request(callback)

	// Получаем готовые к отправке уведомления
	pendingNotifications, err := b.notificationService.GetPendingNotifications()
	if err != nil {
		b.sendMessage(callbackQuery.Message.Chat.ID, "❌ Ошибка получения уведомлений: "+err.Error())
		return
	}

	if len(pendingNotifications) == 0 {
		b.sendMessage(callbackQuery.Message.Chat.ID, "📭 Нет уведомлений для отправки")
		return
	}

	sent := 0
	for _, notification := range pendingNotifications {
		if err := b.sendNotificationToUser(&notification); err != nil {
			log.Printf("Failed to send notification %d: %v", notification.ID, err)
			continue
		}

		// Помечаем как отправленное
		if err := b.notificationService.MarkNotificationSent(notification.ID); err != nil {
			log.Printf("Failed to mark notification %d as sent: %v", notification.ID, err)
		}
		sent++
	}

	b.sendMessage(callbackQuery.Message.Chat.ID,
		fmt.Sprintf("✅ Отправлено уведомлений: %d из %d", sent, len(pendingNotifications)))
}

func (b *Bot) createTestNotification(callbackQuery *tgbotapi.CallbackQuery) {
	// Отвечаем на callback
	callback := tgbotapi.NewCallback(callbackQuery.ID, "Создаю тестовое уведомление...")
	b.api.Request(callback)

	err := b.notificationService.CreateTestNotifications(callbackQuery.From.ID)
	if err != nil {
		b.sendMessage(callbackQuery.Message.Chat.ID, "❌ Ошибка создания тестового уведомления: "+err.Error())
		return
	}

	b.sendMessage(callbackQuery.Message.Chat.ID, "✅ Тестовое уведомление создано! Используйте 'Проверить уведомления' для отправки.")
}

func (b *Bot) sendNotificationToUser(notification *dto.NotificationResponse) error {
	// Получаем информацию о пользователе
	user, err := b.userService.GetUserByID(notification.User.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Получаем информацию о задаче
	task, err := b.taskService.GetTaskByID(notification.Task.ID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Формируем текст уведомления
	var typeText string
	switch notification.Type {
	case "1day":
		typeText = "завтра"
	case "3days":
		typeText = "через 3 дня"
	case "1week":
		typeText = "через неделю"
	default:
		typeText = "скоро"
	}

	text := fmt.Sprintf("🔔 Напоминание о дедлайне!\n\n"+
		"📋 Задача: %s\n"+
		"⏰ Дедлайн %s (%s)\n"+
		"📝 Описание: %s",
		task.Title,
		typeText,
		utils.FormatDeadline(task.Deadline),
		task.Description)

	// Отправляем уведомление
	msg := tgbotapi.NewMessage(user.TelegramID, text)
	_, err = b.api.Send(msg)

	return err
}

func (b *Bot) showMyNotifications(callbackQuery *tgbotapi.CallbackQuery) {
	// Отвечаем на callback
	callback := tgbotapi.NewCallback(callbackQuery.ID, "Получаю ваши уведомления...")
	b.api.Request(callback)

	notifications, err := b.notificationService.GetUserNotifications(callbackQuery.From.ID)
	if err != nil {
		b.sendMessage(callbackQuery.Message.Chat.ID, "❌ Ошибка получения уведомлений: "+err.Error())
		return
	}

	if len(notifications) == 0 {
		b.sendMessage(callbackQuery.Message.Chat.ID, "📭 У вас нет ожидающих уведомлений")
		return
	}

	text := fmt.Sprintf("📋 Ваши ожидающие уведомления (%d):\n\n", len(notifications))
	for i, notif := range notifications {
		text += fmt.Sprintf("%d. %s\n", i+1, notif.Type)
	}

	b.sendMessage(callbackQuery.Message.Chat.ID, text)
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	b.api.Send(msg)
}

func (b *Bot) sendMessageWithKeyboard(chatID int64, text string, keyboard interface{}) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	switch kb := keyboard.(type) {
	case tgbotapi.InlineKeyboardMarkup:
		msg.ReplyMarkup = kb
	case tgbotapi.ReplyKeyboardMarkup:
		msg.ReplyMarkup = kb
	}

	b.api.Send(msg)
}

func (b *Bot) editMessage(chatID int64, messageID int, text string, keyboard *tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}
	b.api.Send(msg)
}

func (b *Bot) ensureUserExists(from *tgbotapi.User) {
	_, err := b.userService.GetUserByTelegramID(from.ID)
	if err != nil {
		userReq := &dto.UserRequest{
			TelegramID: from.ID,
			Username:   from.UserName,
			FirstName:  from.FirstName,
			LastName:   from.LastName,
		}
		b.userService.CreateUser(userReq)
	}
}

func getUserSession(userID int64) *UserSession {
	return userSessions[userID]
}

func setUserSession(userID int64, session *UserSession) {
	userSessions[userID] = session
}

func clearUserSession(userID int64) {
	delete(userSessions, userID)
}

func updateUserSessionState(userID int64, state string) {
	session := getUserSession(userID)
	if session == nil {
		session = &UserSession{
			Data: make(map[string]interface{}),
		}
	}
	session.State = state
	setUserSession(userID, session)
}

func (b *Bot) setCommands() error {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Начать работу с ботом"},
		{Command: "help", Description: "Справка по командам"},
		{Command: "new_group", Description: "Создать новую группу"},
		{Command: "connect", Description: "Подключиться к группе"},
		{Command: "my_groups", Description: "Мои группы"},
		{Command: "select_group", Description: "Выбрать активную группу"},
		{Command: "new_task", Description: "Создать задачу"},
		{Command: "tasks", Description: "Показать задачи"},
		{Command: "export", Description: "Экспорт задач"},
		{Command: "test_notifications", Description: "Тестировать уведомления"},
		{Command: "create_test_task", Description: "Создать тестовую задачу"},
		{Command: "cancel", Description: "Отменить операцию"},
	}

	setCommandsConfig := tgbotapi.NewSetMyCommands(commands...)
	_, err := b.api.Request(setCommandsConfig)
	return err
}

func (b *Bot) SendNotification(userTelegramID int64, message string) error {
	msg := tgbotapi.NewMessage(userTelegramID, message)
	_, err := b.api.Send(msg)
	if err != nil {
		log.Printf("❌ Failed to send notification to user %d: %v", userTelegramID, err)
	} else {
		log.Printf("✅ Sent notification to user %d", userTelegramID)
	}
	return err
}

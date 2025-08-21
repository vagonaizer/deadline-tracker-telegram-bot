package telegram

import (
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const imagesPath = "assets/images"

func (b *Bot) sendWelcomeImage(chatID int64) {
	imagePath := filepath.Join(imagesPath, "welcome.jpg")
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imagePath))
	photo.Caption = "🚀 Начните работу с DeadlineBot!"
	b.api.Send(photo)
}

func (b *Bot) sendCreateGroupImage(chatID int64) {
	imagePath := filepath.Join(imagesPath, "create_group.jpg")
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imagePath))
	photo.Caption = "📋 Шаги создания новой группы"
	b.api.Send(photo)
}

func (b *Bot) sendConnectGroupImage(chatID int64) {
	imagePath := filepath.Join(imagesPath, "connect_group.jpg")
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imagePath))
	photo.Caption = "🔗 Как подключиться к группе"
	b.api.Send(photo)
}

func (b *Bot) sendCreateTaskImage(chatID int64) {
	imagePath := filepath.Join(imagesPath, "create_task.jpg")
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imagePath))
	photo.Caption = "📝 Создание новой задачи"
	b.api.Send(photo)
}

func (b *Bot) sendPrioritiesImage(chatID int64) {
	imagePath := filepath.Join(imagesPath, "priorities.jpg")
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imagePath))
	photo.Caption = "🎯 Система приоритетов задач"
	b.api.Send(photo)
}

func (b *Bot) sendExportDemoImage(chatID int64) {
	imagePath := filepath.Join(imagesPath, "export_demo.jpg")
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imagePath))
	photo.Caption = "📊 Примеры экспорта данных"
	b.api.Send(photo)
}

func (b *Bot) sendTaskManagementImage(chatID int64) {
	imagePath := filepath.Join(imagesPath, "task_management.jpg")
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imagePath))
	photo.Caption = "⚙️ Управление задачами"
	b.api.Send(photo)
}

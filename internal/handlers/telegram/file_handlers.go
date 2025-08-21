package telegram

import (
	"deadline-bot/internal/dto"
	"deadline-bot/pkg/utils"

	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleDocument(message *tgbotapi.Message) {
	session := getUserSession(message.From.ID)
	if session == nil || session.GroupID == 0 {
		b.sendMessage(message.Chat.ID, "Сначала выберите группу командой /select_group")
		return
	}

	if session.State != "uploading_file" {
		b.sendMessage(message.Chat.ID, "Отправьте команду /upload_file для загрузки файла к задаче")
		return
	}

	taskID, ok := session.Data["task_id"].(uint)
	if !ok {
		b.sendMessage(message.Chat.ID, "Ошибка: не указана задача для загрузки файла")
		clearUserSession(message.From.ID)
		return
	}

	document := message.Document
	if document == nil {
		b.sendMessage(message.Chat.ID, "Ошибка: документ не найден")
		return
	}

	fileConfig := tgbotapi.FileConfig{FileID: document.FileID}
	file, err := b.api.GetFile(fileConfig)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при получении файла: "+err.Error())
		return
	}

	fileURL := file.Link(b.api.Token)
	resp, err := http.Get(fileURL)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при скачивании файла: "+err.Error())
		return
	}
	defer resp.Body.Close()

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при чтении файла: "+err.Error())
		return
	}

	mimeType := document.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	uploadReq := &dto.FileUploadRequest{
		TaskID:   taskID,
		FileName: document.FileName,
		FileData: fileData,
		MimeType: mimeType,
	}

	uploadedFile, err := b.fileService.UploadFile(message.From.ID, uploadReq)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при загрузке файла: "+err.Error())
		return
	}

	clearUserSession(message.From.ID)
	setUserSession(message.From.ID, &UserSession{GroupID: session.GroupID, Data: make(map[string]interface{})})

	text := fmt.Sprintf(`✅ *Файл успешно загружен!*

📎 *Название:* %s
📊 *Размер:* %s
📄 *Тип:* %s`,
		utils.EscapeMarkdown(uploadedFile.FileName),
		utils.FormatFileSize(uploadedFile.FileSize),
		uploadedFile.MimeType)

	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) handleUploadFileCommand(message *tgbotapi.Message) {
	session := getUserSession(message.From.ID)
	if session == nil || session.GroupID == 0 {
		b.sendMessage(message.Chat.ID, "Сначала выберите группу командой /select_group")
		return
	}

	args := strings.Fields(message.CommandArguments())
	if len(args) != 1 {
		b.sendMessage(message.Chat.ID, "Использование: /upload_file <ID_задачи>")
		return
	}

	taskID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Неверный ID задачи")
		return
	}

	task, err := b.taskService.GetTaskByID(uint(taskID))
	if err != nil {
		b.sendMessage(message.Chat.ID, "Задача не найдена: "+err.Error())
		return
	}

	if task.Group.ID != session.GroupID {
		b.sendMessage(message.Chat.ID, "Задача не принадлежит выбранной группе")
		return
	}

	session.State = "uploading_file"
	session.Data["task_id"] = uint(taskID)
	setUserSession(message.From.ID, session)

	text := fmt.Sprintf(`📎 *Загрузка файла к задаче*

📋 *Задача:* %s

Отправьте файл (документ) для загрузки.
Поддерживаемые форматы: PDF, DOC, DOCX, ZIP, RAR, TXT, JPG, PNG

Для отмены используйте /cancel`,
		utils.EscapeMarkdown(task.Title))

	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) handleTaskFilesCommand(message *tgbotapi.Message) {
	args := strings.Fields(message.CommandArguments())
	if len(args) != 1 {
		b.sendMessage(message.Chat.ID, "Использование: /task_files <ID_задачи>")
		return
	}

	taskID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Неверный ID задачи")
		return
	}

	task, err := b.taskService.GetTaskByID(uint(taskID))
	if err != nil {
		b.sendMessage(message.Chat.ID, "Задача не найдена: "+err.Error())
		return
	}

	files, err := b.fileService.GetTaskFiles(uint(taskID))
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при получении файлов: "+err.Error())
		return
	}

	if len(files) == 0 {
		text := fmt.Sprintf(`📎 *Файлы задачи "%s"*

К этой задаче пока не прикреплены файлы.

Используйте /upload_file %d для загрузки файла.`,
			utils.EscapeMarkdown(task.Title), taskID)

		b.sendMessage(message.Chat.ID, text)
		return
	}

	var text strings.Builder
	text.WriteString(fmt.Sprintf("📎 *Файлы задачи \"%s\"*\n\n", utils.EscapeMarkdown(task.Title)))

	for i, file := range files {
		text.WriteString(fmt.Sprintf("%d. 📄 *%s*\n", i+1, utils.EscapeMarkdown(file.FileName)))
		text.WriteString(fmt.Sprintf("   📊 %s • %s\n",
			utils.FormatFileSize(file.FileSize),
			file.MimeType))

		if i < len(files)-1 {
			text.WriteString("\n")
		}
	}

	b.sendMessage(message.Chat.ID, text.String())
}

func (b *Bot) handleDownloadFileCommand(message *tgbotapi.Message) {
	args := strings.Fields(message.CommandArguments())
	if len(args) != 1 {
		b.sendMessage(message.Chat.ID, "Использование: /download_file <ID_файла>")
		return
	}

	fileID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Неверный ID файла")
		return
	}

	file, err := b.fileService.GetFileByID(uint(fileID))
	if err != nil {
		b.sendMessage(message.Chat.ID, "Файл не найден: "+err.Error())
		return
	}

	filePath, err := b.fileService.GetFilePath(uint(fileID))
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при получении файла: "+err.Error())
		return
	}

	document := tgbotapi.NewDocument(message.Chat.ID, tgbotapi.FilePath(filePath))
	document.Caption = fmt.Sprintf("📎 %s\n📊 %s",
		file.FileName,
		utils.FormatFileSize(file.FileSize))

	if _, err := b.api.Send(document); err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при отправке файла: "+err.Error())
		return
	}
}

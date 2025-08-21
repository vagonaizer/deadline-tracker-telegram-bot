package utils

import (
	"deadline-bot/internal/dto"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

func GeneratePDF(tasks []dto.ExportTaskData, groupName string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Устанавливаем шрифт (поддерживающий кириллицу через транслитерацию)
	pdf.SetFont("Arial", "B", 16)

	// Заголовок
	title := fmt.Sprintf("Zadachi gruppy: %s", transliterate(groupName))
	pdf.Cell(0, 10, title)
	pdf.Ln(15)

	// Заголовки таблицы
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(60, 8, "Nazvanie")
	pdf.Cell(40, 8, "Dedlajn")
	pdf.Cell(25, 8, "Prioritet")
	pdf.Cell(25, 8, "Status")
	pdf.Ln(10)

	// Линия
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(2)

	// Данные
	pdf.SetFont("Arial", "", 9)
	for _, task := range tasks {
		if pdf.GetY() > 270 { // Если страница заканчивается
			pdf.AddPage()
		}

		// Название задачи (транслитерация)
		taskTitle := transliterate(task.Title)
		if len(taskTitle) > 35 {
			taskTitle = taskTitle[:32] + "..."
		}
		pdf.Cell(60, 6, taskTitle)

		// Дедлайн
		deadline := task.Deadline.Format("02.01.2006 15:04")
		pdf.Cell(40, 6, deadline)

		// Приоритет
		priority := translatePriority(task.Priority)
		pdf.Cell(25, 6, priority)

		// Статус
		status := translateStatus(task.Status)
		pdf.Cell(25, 6, status)

		pdf.Ln(8)

		// Описание (если есть)
		if task.Description != "" {
			pdf.SetFont("Arial", "I", 8)
			description := transliterate(task.Description)
			if len(description) > 80 {
				description = description[:77] + "..."
			}
			pdf.Cell(10, 4, "")
			pdf.Cell(0, 4, description)
			pdf.Ln(6)
			pdf.SetFont("Arial", "", 9)
		}
	}

	// Футер
	pdf.Ln(10)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(5)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 4, fmt.Sprintf("Vygruženo: %s", FormatDeadline(time.Now())))

	// Возвращаем PDF как байты
	var buf []byte
	var err error

	// Создаем буфер для записи PDF
	w := &bytesWriter{buf: &buf}
	err = pdf.Output(w)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf, nil
}

type bytesWriter struct {
	buf *[]byte
}

func (w *bytesWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

// Простая транслитерация кириллицы
func transliterate(text string) string {
	translitMap := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo",
		'ж': "zh", 'з': "z", 'и': "i", 'й': "j", 'к': "k", 'л': "l", 'м': "m",
		'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
		'ф': "f", 'х': "h", 'ц': "c", 'ч': "ch", 'ш': "sh", 'щ': "sch",
		'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
		'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D", 'Е': "E", 'Ё': "Yo",
		'Ж': "Zh", 'З': "Z", 'И': "I", 'Й': "J", 'К': "K", 'Л': "L", 'М': "M",
		'Н': "N", 'О': "O", 'П': "P", 'Р': "R", 'С': "S", 'Т': "T", 'У': "U",
		'Ф': "F", 'Х': "H", 'Ц': "C", 'Ч': "Ch", 'Ш': "Sh", 'Щ': "Sch",
		'Ъ': "", 'Ы': "Y", 'Ь': "", 'Э': "E", 'Ю': "Yu", 'Я': "Ya",
	}

	result := ""
	for _, char := range text {
		if replacement, exists := translitMap[char]; exists {
			result += replacement
		} else {
			result += string(char)
		}
	}
	return result
}

func translatePriority(priority string) string {
	switch priority {
	case "low":
		return "Nizkij"
	case "normal":
		return "Obychnyj"
	case "high":
		return "Vysokij"
	case "critical":
		return "Kriticheskij"
	default:
		return priority
	}
}

func translateStatus(status string) string {
	switch status {
	case "pending":
		return "V rabote"
	case "completed":
		return "Vypolneno"
	case "overdue":
		return "Prosrocheno"
	default:
		return status
	}
}

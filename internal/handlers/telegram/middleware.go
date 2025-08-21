package telegram

import (
	"deadline-bot/internal/services"
	"log"
	"runtime/debug"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Middleware struct {
	userService services.UserService
}

func NewMiddleware(userService services.UserService) *Middleware {
	return &Middleware{
		userService: userService,
	}
}

func (m *Middleware) RecoveryMiddleware(handler func(update tgbotapi.Update)) func(update tgbotapi.Update) {
	return func(update tgbotapi.Update) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered in telegram handler: %v\n%s", err, debug.Stack())
			}
		}()

		handler(update)
	}
}

func (m *Middleware) LoggingMiddleware(handler func(update tgbotapi.Update)) func(update tgbotapi.Update) {
	return func(update tgbotapi.Update) {
		start := time.Now()

		var userID int64
		var userName string
		var command string

		if update.Message != nil {
			userID = update.Message.From.ID
			userName = update.Message.From.UserName
			if update.Message.IsCommand() {
				command = update.Message.Command()
			}
		} else if update.CallbackQuery != nil {
			userID = update.CallbackQuery.From.ID
			userName = update.CallbackQuery.From.UserName
			command = "callback:" + update.CallbackQuery.Data
		}

		log.Printf("Telegram update: user_id=%d user_name=%s command=%s", userID, userName, command)

		handler(update)

		duration := time.Since(start)
		log.Printf("Telegram update processed: user_id=%d duration=%v", userID, duration)
	}
}

func (m *Middleware) RateLimitMiddleware(handler func(update tgbotapi.Update)) func(update tgbotapi.Update) {
	userLastRequest := make(map[int64]time.Time)
	rateLimitDuration := 1 * time.Second

	return func(update tgbotapi.Update) {
		var userID int64

		if update.Message != nil {
			userID = update.Message.From.ID
		} else if update.CallbackQuery != nil {
			userID = update.CallbackQuery.From.ID
		} else {
			handler(update)
			return
		}

		now := time.Now()
		lastRequest, exists := userLastRequest[userID]

		if exists && now.Sub(lastRequest) < rateLimitDuration {
			log.Printf("Rate limit exceeded for user %d", userID)
			return
		}

		userLastRequest[userID] = now
		handler(update)
	}
}

func (m *Middleware) AuthMiddleware(handler func(update tgbotapi.Update)) func(update tgbotapi.Update) {
	return func(update tgbotapi.Update) {
		var user *tgbotapi.User

		if update.Message != nil {
			user = update.Message.From
		} else if update.CallbackQuery != nil {
			user = update.CallbackQuery.From
		} else {
			handler(update)
			return
		}

		if user.IsBot {
			log.Printf("Ignoring bot user: %d", user.ID)
			return
		}

		handler(update)
	}
}

func ChainMiddleware(handler func(update tgbotapi.Update), middlewares ...func(func(update tgbotapi.Update)) func(update tgbotapi.Update)) func(update tgbotapi.Update) {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

package app

import (
	"context"
	"deadline-bot/internal/config"
	"deadline-bot/internal/repositories"
	"deadline-bot/internal/repositories/storage"
	"deadline-bot/internal/services"

	httpHandlers "deadline-bot/internal/handlers/http"
	"deadline-bot/internal/handlers/telegram"
	"deadline-bot/pkg/database/postgres"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	config      *config.Config
	database    *postgres.Database
	httpServer  *http.Server
	telegramBot *telegram.Bot
	scheduler   services.SchedulerService
}

func NewApp(configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	repositories := initRepositories(db)
	services := initServices(repositories, cfg)

	telegramBot, err := initTelegramBot(services, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init telegram bot: %w", err)
	}

	httpServer := initHTTPServer(services, db, cfg)

	return &App{
		config:      cfg,
		database:    db,
		httpServer:  httpServer,
		telegramBot: telegramBot,
		scheduler:   services.Scheduler,
	}, nil
}

func (a *App) Run() error {
	log.Println("Starting application...")

	if err := a.scheduler.StartScheduler(); err != nil {
		log.Printf("Warning: failed to start scheduler: %v", err)
	}

	go func() {
		log.Printf("Starting HTTP server on %s:%d", a.config.Server.Host, a.config.Server.Port)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	go func() {
		log.Println("Starting Telegram bot...")
		if err := a.telegramBot.Start(); err != nil {
			log.Printf("Telegram bot error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down application...")
	return a.Shutdown()
}

func (a *App) Shutdown() error {
	log.Println("Stopping scheduler...")
	if err := a.scheduler.StopScheduler(); err != nil {
		log.Printf("Error stopping scheduler: %v", err)
	}

	log.Println("Stopping Telegram bot...")
	a.telegramBot.Stop()

	log.Println("Stopping HTTP server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Error stopping HTTP server: %v", err)
	}

	log.Println("Closing database connection...")
	if err := a.database.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Application stopped")
	return nil
}

type Repositories struct {
	User         repositories.UserRepository
	Group        repositories.GroupRepository
	UserGroup    repositories.UserGroupRepository
	Task         repositories.TaskRepository
	File         repositories.FileRepository
	Notification repositories.NotificationRepository
}

type Services struct {
	User         services.UserService
	Group        services.GroupService
	Task         services.TaskService
	Auth         services.AuthService
	File         services.FileService
	Notification services.NotificationService
	Export       services.ExportService
	Scheduler    services.SchedulerService
}

func initRepositories(db *postgres.Database) *Repositories {
	gormDB := db.GetDB()

	return &Repositories{
		User:         storage.NewUserRepository(gormDB),
		Group:        storage.NewGroupRepository(gormDB),
		UserGroup:    storage.NewUserGroupRepository(gormDB),
		Task:         storage.NewTaskRepository(gormDB),
		File:         storage.NewFileRepository(gormDB),
		Notification: storage.NewNotificationRepository(gormDB),
	}
}

func initServices(repos *Repositories, cfg *config.Config) *Services {
	// Отладочный вывод конфига
	log.Printf("DEBUG: Notifications config - EnableReminders: %v, CheckIntervalHours: %f",
		cfg.Notifications.EnableReminders, cfg.Notifications.CheckIntervalHours)

	authService := services.NewAuthService(repos.User, repos.Group, repos.UserGroup)
	userService := services.NewUserService(repos.User)
	groupService := services.NewGroupService(repos.User, repos.Group, repos.UserGroup, authService)

	// Создаем NotificationService без зависимостей от внешних API
	notificationService := services.NewNotificationService(
		repos.Notification, repos.Task, repos.User, repos.UserGroup, &cfg.Notifications)

	taskService := services.NewTaskService(repos.User, repos.Task, repos.UserGroup, authService, notificationService)
	fileService := services.NewFileService(
		repos.File, repos.Task, repos.User, repos.UserGroup, authService, &cfg.Files)
	exportService := services.NewExportService(repos.Task, repos.User, repos.UserGroup, authService)
	schedulerService := services.NewSchedulerService(
		repos.Task, repos.Group, notificationService, &cfg.Notifications)

	return &Services{
		User:         userService,
		Group:        groupService,
		Task:         taskService,
		Auth:         authService,
		File:         fileService,
		Notification: notificationService,
		Export:       exportService,
		Scheduler:    schedulerService,
	}
}

func initTelegramBot(services *Services, cfg *config.Config) (*telegram.Bot, error) {
	bot, err := telegram.NewBot(
		&cfg.Bot,
		services.User,
		services.Group,
		services.Task,
		services.Auth,
		services.File,
		services.Export,
		services.Notification, // добавляем NotificationService
	)
	if err != nil {
		return nil, err
	}

	return bot, nil
}

func initHTTPServer(services *Services, db *postgres.Database, cfg *config.Config) *http.Server {
	if !cfg.Bot.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	healthHandler := httpHandlers.NewHealthHandler(db.GetDB())
	healthHandler.RegisterRoutes(router)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	return &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

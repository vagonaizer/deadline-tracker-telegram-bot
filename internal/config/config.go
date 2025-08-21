package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Bot           BotConfig           `mapstructure:"bot"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Server        ServerConfig        `mapstructure:"server"`
	Files         FilesConfig         `mapstructure:"files"`
	Notifications NotificationsConfig `mapstructure:"notifications"`
}

type BotConfig struct {
	Token   string `mapstructure:"token"`
	Debug   bool   `mapstructure:"debug"`
	Timeout int    `mapstructure:"timeout"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type FilesConfig struct {
	UploadPath  string   `mapstructure:"upload_path"`
	MaxSizeMB   int      `mapstructure:"max_size_mb"`
	AllowedExts []string `mapstructure:"allowed_extensions"`
}

type NotificationsConfig struct {
	EnableReminders    bool    `mapstructure:"enable_reminders"`
	CheckIntervalHours float64 `mapstructure:"check_interval_hours"` // изменили на float64
	DefaultReminders   []int   `mapstructure:"default_reminders"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)

	viper.SetDefault("bot.timeout", 60)
	viper.SetDefault("bot.debug", false)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("files.upload_path", "./uploads")
	viper.SetDefault("files.max_size_mb", 50)
	viper.SetDefault("files.allowed_extensions", []string{".pdf", ".zip", ".rar", ".txt", ".doc", ".docx"})
	viper.SetDefault("notifications.enable_reminders", true)
	viper.SetDefault("notifications.check_interval_hours", 6.0) // изменили на float64
	viper.SetDefault("notifications.default_reminders", []int{1, 3, 7})

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func validateConfig(config *Config) error {
	if config.Bot.Token == "" {
		return fmt.Errorf("bot token is required")
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if config.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	if config.Files.MaxSizeMB <= 0 {
		return fmt.Errorf("files max size must be positive")
	}

	if config.Notifications.CheckIntervalHours <= 0 {
		return fmt.Errorf("notification check interval must be positive")
	}

	return nil
}

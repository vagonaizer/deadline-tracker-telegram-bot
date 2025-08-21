package models

import "time"

const (
	NotificationTypeOneDay    = "1day"
	NotificationTypeThreeDays = "3days"
	NotificationTypeOneWeek   = "1week"
)

type Notification struct {
	ID            uint       `gorm:"primaryKey"`
	UserID        uint       `gorm:"not null;index"`
	TaskID        uint       `gorm:"not null;index"`
	Type          string     `gorm:"not null;index"`
	ScheduledTime *time.Time `gorm:"index"` // Время когда нужно отправить уведомление
	SentAt        *time.Time // Время когда уведомление было отправлено
	CreatedAt     time.Time
	UpdatedAt     time.Time

	User User `gorm:"foreignKey:UserID"`
	Task Task `gorm:"foreignKey:TaskID"`
}

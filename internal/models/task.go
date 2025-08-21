package models

import "time"

type Task struct {
	ID          uint      `gorm:"primaryKey"`
	Title       string    `gorm:"size:200;not null"`
	Description string    `gorm:"type:text"`
	Deadline    time.Time `gorm:"not null"`
	Priority    string    `gorm:"size:20;not null;default:'normal'"`
	Status      string    `gorm:"size:20;not null;default:'pending'"`
	GroupID     uint      `gorm:"not null"`
	CreatedBy   uint      `gorm:"not null"`
	AssignedTo  *uint     `gorm:"null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Group         Group          `gorm:"foreignKey:GroupID"`
	Creator       User           `gorm:"foreignKey:CreatedBy"`
	Assignee      *User          `gorm:"foreignKey:AssignedTo"`
	Files         []File         `gorm:"foreignKey:TaskID"`
	Notifications []Notification `gorm:"foreignKey:TaskID"`
}

const (
	PriorityLow      = "low"
	PriorityNormal   = "normal"
	PriorityHigh     = "high"
	PriorityCritical = "critical"

	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusOverdue   = "overdue"
)

package models

import "time"

type User struct {
	ID         uint   `gorm:"primaryKey"`
	TelegramID int64  `gorm:"uniqueIndex;not null"`
	Username   string `gorm:"size:100"`
	FirstName  string `gorm:"size:100"`
	LastName   string `gorm:"size:100"`
	CreatedAt  time.Time
	UpdatedAt  time.Time

	UserGroups    []UserGroup    `gorm:"foreignKey:UserID"`
	CreatedGroups []Group        `gorm:"foreignKey:CreatedBy"`
	AssignedTasks []Task         `gorm:"foreignKey:AssignedTo"`
	Notifications []Notification `gorm:"foreignKey:UserID"`
}

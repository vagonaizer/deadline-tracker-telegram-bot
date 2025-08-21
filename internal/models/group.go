package models

import "time"

type Group struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:200;not null"`
	Login     string `gorm:"size:50;uniqueIndex;not null"`
	Password  string `gorm:"size:255;not null"`
	CreatedBy uint   `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Creator    User        `gorm:"foreignKey:CreatedBy"`
	UserGroups []UserGroup `gorm:"foreignKey:GroupID"`
	Tasks      []Task      `gorm:"foreignKey:GroupID"`
}

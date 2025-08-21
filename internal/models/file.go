package models

import "time"

type File struct {
	ID        uint   `gorm:"primaryKey"`
	TaskID    uint   `gorm:"not null"`
	FileName  string `gorm:"size:255;not null"`
	FilePath  string `gorm:"size:500;not null"`
	FileSize  int64  `gorm:"not null"`
	MimeType  string `gorm:"size:100;not null"`
	CreatedAt time.Time

	Task Task `gorm:"foreignKey:TaskID"`
}

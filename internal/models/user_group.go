package models

import "time"

type UserGroup struct {
	UserID   uint      `gorm:"primaryKey"`
	GroupID  uint      `gorm:"primaryKey"`
	Role     string    `gorm:"size:20;not null;default:'member'"`
	JoinedAt time.Time `gorm:"not null"`

	User  User  `gorm:"foreignKey:UserID"`
	Group Group `gorm:"foreignKey:GroupID"`
}

const (
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
	RoleMember    = "member"
)

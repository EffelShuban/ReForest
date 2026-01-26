package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct{
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	RoleType string `gorm:"type:varchar(20)"`
	CreateAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Admin struct{
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`
	EmployeeID string `gorm:"unique"`
	AccessLevel string
}

type Sponsor struct{
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Balance int `gorm:"default:0"`
}

type AuthClaims struct{
	UserID uuid.UUID
	Role string
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct{
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"unique;not null"`
	PasswordHash string    `gorm:"not null"`
	RoleType     string    `gorm:"type:varchar(20)"`
	Profile      Profile   `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt    time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Profile struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	FullName    string    `gorm:"not null"`
	DateOfBirth time.Time
	Age         int
	Balance     int `gorm:"default:0"`
}

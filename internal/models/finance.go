package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wallet struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	Balance   int64     `gorm:"default:0"` // Stored in smallest unit (e.g., cents)
	UpdatedAt time.Time
}

type Transaction struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;index;not null"`
	Amount      int64     `gorm:"not null"`
	Type        string    `gorm:"type:varchar(20);not null"` // DEPOSIT, WITHDRAWAL, PURCHASE
	Status      string    `gorm:"type:varchar(20);default:'PENDING'"` // PENDING, COMPLETED, FAILED, EXPIRED
	ReferenceID string    `gorm:"type:varchar(100)"`          // External Payment ID (e.g. Xendit Invoice ID)
	PaymentURL  string    `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
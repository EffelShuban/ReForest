package models

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TransactionID uuid.UUID `gorm:"type:uuid;not null"`
	Amount        int64     `gorm:"not null"`
	Status        string    `gorm:"default:'PENDING'"` // PENDING, FINISHED, FAILED, EXPIRED
	ExternalID    string
	PaymentURL    string
	ExpiresAt     time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Transaction struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Amount    int64     `gorm:"not null"`
	Type      string    `gorm:"not null"` // DEPOSIT, ADOPT, CARE
	Payment   Payment   `gorm:"foreignKey:TransactionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

package models

import "time"

type Transaction struct {
	ID              string    `bson:"_id,omitempty"`
	SponsorID       string    `bson:"sponsor_id,omitempty"`
	AdoptedTreeID   string    `bson:"adopted_tree_id,omitempty"`
	Amount          float64   `bson:"amount,omitempty"`
	Type            string    `bson:"type,omitempty"`
	PaymentStatus   string    `bson:"payment_status,omitempty"`
	TransactionDate time.Time `bson:"transaction_date,omitempty"`
}

type Payment struct {
	ID            string    `bson:"_id,omitempty"`
	TransactionID string    `bson:"transaction_id,omitempty"`
	Amount        float64   `bson:"amount,omitempty"`
	Status        string    `bson:"status,omitempty"`
	CreatedAt     time.Time `bson:"created_at,omitempty"`
}

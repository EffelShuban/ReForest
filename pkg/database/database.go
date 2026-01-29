package database

import (
	"log"

	"reforest/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewConnection establishes a connection to the database and runs migrations.
func NewConnection(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	log.Println("Running migrations...")
	if err := db.AutoMigrate(&models.User{}, &models.Profile{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	return db
}
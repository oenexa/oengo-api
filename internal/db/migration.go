package db

import (
	"log"

	"oengo-api/internal/models"
	"gorm.io/gorm"
)

// AutoMigrateBankingSchema sets up the professional PostgreSQL tables
// using Banking Standard constraints and Decimal precision.
func AutoMigrateBankingSchema(db *gorm.DB) error {
	log.Println("Starting Banking Standard Database Schema Migration...")

	// Execute migrations for all strict models
	err := db.AutoMigrate(
		&models.BankingUser{},
		&models.BankingWallet{},
		&models.LedgerAccount{},
		&models.LedgerTransaction{},
		&models.LedgerEntryRecord{},
		&models.BankingOrder{},
	)

	if err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}

	log.Println("Database Schema Migration completed successfully.")
	return nil
}

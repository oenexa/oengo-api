package repository

import (
	"errors"
	"fmt"

	"oengo-api/internal/models"
	"gorm.io/gorm"
)

type PostgresWalletRepository struct {
	db *gorm.DB
}

func NewPostgresWalletRepository(db *gorm.DB) *PostgresWalletRepository {
	db.AutoMigrate(&models.User{}, &models.LedgerEntry{})
	return &PostgresWalletRepository{db: db}
}

func (r *PostgresWalletRepository) GetWalletByUserID(userID string) (*models.DigitalWallet, error) {
	var user models.User
	result := r.db.Where("id = ?", userID).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user.DigitalWallet, nil
}

func (r *PostgresWalletRepository) UpdateWallet(userID string, wallet *models.DigitalWallet) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"fiat_balance_eur":   wallet.FiatBalanceEUR,
		"locked_balance_eur": wallet.LockedBalanceEUR,
		"loyalty_points":    wallet.LoyaltyPoints,
		"cashback_rate_pct":  wallet.CashbackRatePct,
	}).Error
}

func (r *PostgresWalletRepository) Transact(entries []models.LedgerEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Verify ledger balancing constraint before committing transaction
	var debitTotal, creditTotal float64
	for _, e := range entries {
		if e.EntryType == models.LedgerDebit {
			debitTotal += e.AmountEUR
		} else if e.EntryType == models.LedgerCredit {
			creditTotal += e.AmountEUR
		}
	}

	// In double-entry bookkeeping, debits must equal credits
	diff := debitTotal - creditTotal
	if diff < -0.001 || diff > 0.001 {
		return fmt.Errorf("double-entry invariant violated: debits (%.2f) != credits (%.2f)", debitTotal, creditTotal)
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, entry := range entries {
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PostgresWalletRepository) GetLedgerEntries(limit int) ([]models.LedgerEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	var entries []models.LedgerEntry
	err := r.db.Order("created_at desc").Limit(limit).Find(&entries).Error
	return entries, err
}

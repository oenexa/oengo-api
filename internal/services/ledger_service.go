package services

import (
	"errors"

	"oengo-api/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// RecordTransaction strictly enforces Double-Entry bookkeeping.
// SUM(DEBIT) must exactly equal SUM(CREDIT).
func RecordTransaction(tx *gorm.DB, ledgerTx *models.LedgerTransaction) error {
	debits := decimal.NewFromInt(0)
	credits := decimal.NewFromInt(0)

	// 1. Calculate and verify totals
	for _, entry := range ledgerTx.Entries {
		if entry.Amount.LessThanOrEqual(decimal.Zero) {
			return errors.New("entry amounts must be strictly positive")
		}
		if entry.Direction == models.DirectionDebit {
			debits = debits.Add(entry.Amount)
		} else if entry.Direction == models.DirectionCredit {
			credits = credits.Add(entry.Amount)
		} else {
			return errors.New("invalid entry direction")
		}
	}

	if !debits.Equal(credits) {
		return errors.New("transaction unbalanced: debits must equal credits")
	}

	// 2. Persist Immutable Transaction & Entries
	if err := tx.Create(ledgerTx).Error; err != nil {
		return err
	}

	// 3. Update Cached Wallet Balances
	for _, entry := range ledgerTx.Entries {
		// Only update wallet if the account belongs to a user wallet
		// NOTE: In a full system, you would check if AccountID maps to a Wallet
		var wallet models.BankingWallet
		if err := tx.Where("account_id = ?", entry.AccountID).First(&wallet).Error; err == nil {
			if entry.Direction == models.DirectionCredit {
				wallet.Balance = wallet.Balance.Add(entry.Amount)
			} else {
				wallet.Balance = wallet.Balance.Sub(entry.Amount)
			}
			if err := tx.Save(&wallet).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

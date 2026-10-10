package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"oengo-api/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// RecordTransaction strictly enforces Double-Entry bookkeeping, 
// generates cryptographic hashchains, and executes atomic balance updates.
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

	// 2. Hashchain Generation (Cryptographic Ledger)
	var lastTx models.LedgerTransaction
	if err := tx.Order("created_at desc").First(&lastTx).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	
	if lastTx.Hash == "" {
		ledgerTx.PreviousHash = "GENESIS"
	} else {
		ledgerTx.PreviousHash = lastTx.Hash
	}

	hashData := fmt.Sprintf("%s|%s|%s|%d", ledgerTx.PreviousHash, ledgerTx.IdempotencyKey, ledgerTx.ReferenceID.String(), time.Now().UnixNano())
	hash := sha256.Sum256([]byte(hashData))
	ledgerTx.Hash = hex.EncodeToString(hash[:])

	// 3. Persist Immutable Transaction & Entries
	if err := tx.Create(ledgerTx).Error; err != nil {
		return err
	}

	// 4. ATOMIC Update Cached Wallet Balances (Prevents Concurrency Lost Updates)
	for _, entry := range ledgerTx.Entries {
		var expr interface{}
		if entry.Direction == models.DirectionCredit {
			expr = gorm.Expr("balance + ?", entry.Amount)
		} else {
			expr = gorm.Expr("balance - ?", entry.Amount)
		}
		
		// Update exactly the wallet row without loading it into memory (Race-Condition Proof)
		res := tx.Model(&models.BankingWallet{}).Where("account_id = ?", entry.AccountID).UpdateColumn("balance", expr)
		if res.Error != nil {
			return res.Error
		}
	}

	return nil
}

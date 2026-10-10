package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"oengo-api/internal/config"
	"oengo-api/internal/models"
	"gorm.io/gorm"
)

type Store struct {
	mu               sync.RWMutex
	DB               *gorm.DB
	cfg              *config.Config
	Users            map[string]*models.User
	CustomerProfiles map[string]*models.CustomerProfile
	Restaurants      map[string]*models.Restaurant
	Menus            map[string][]models.MenuItem
	Orders           map[string]*models.Order
	Ledger           []models.LedgerEntry
	Coins            map[string]*models.CoinProfile
	Riders           map[string]*models.RiderProfile
	Coupons          map[string]*models.Coupon
	SavedCards       map[string][]models.CardPaymentInfo
	CommissionPct    float64
}

func Sha256Hash(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}

func NewStore(cfg *config.Config, db *gorm.DB) *Store {
	s := &Store{
		cfg:              cfg,
		DB:               db,
		Users:            make(map[string]*models.User),
		CustomerProfiles: make(map[string]*models.CustomerProfile),
		Restaurants:      make(map[string]*models.Restaurant),
		Menus:            make(map[string][]models.MenuItem),
		Orders:           make(map[string]*models.Order),
		Ledger:           make([]models.LedgerEntry, 0),
		Coins:            make(map[string]*models.CoinProfile),
		Riders:           make(map[string]*models.RiderProfile),
		Coupons:          make(map[string]*models.Coupon),
		SavedCards:       make(map[string][]models.CardPaymentInfo),
		CommissionPct:    cfg.DefaultCommissionPct,
	}
	s.seedData()
	return s
}

func (s *Store) seedData() {
	// Seed data removed as requested by user. Starting from scratch.
}

func (s *Store) RecordLedgerTx(txID, debitAcc, creditAcc string, amountEUR float64, desc string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if amountEUR <= 0 {
		return fmt.Errorf("ledger amount must be greater than zero")
	}

	ts := time.Now().Format(time.RFC3339)
	debit := models.LedgerEntry{
		ID:          fmt.Sprintf("ledg_%d_d", time.Now().UnixNano()),
		TxID:        txID,
		AccountCode: debitAcc,
		EntryType:   models.LedgerDebit,
		AmountEUR:   amountEUR,
		Description: desc,
		CreatedAt:   ts,
	}
	credit := models.LedgerEntry{
		ID:          fmt.Sprintf("ledg_%d_c", time.Now().UnixNano()),
		TxID:        txID,
		AccountCode: creditAcc,
		EntryType:   models.LedgerCredit,
		AmountEUR:   amountEUR,
		Description: desc,
		CreatedAt:   ts,
	}

	s.Ledger = append(s.Ledger, debit, credit)
	return nil
}

func (s *Store) Lock()    { s.mu.Lock() }
func (s *Store) Unlock()  { s.mu.Unlock() }
func (s *Store) RLock()   { s.mu.RLock() }
func (s *Store) RUnlock() { s.mu.RUnlock() }

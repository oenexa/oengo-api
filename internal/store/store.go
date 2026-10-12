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

func NewStore(cfg *config.Config, dbs ...*gorm.DB) *Store {
	var db *gorm.DB
	if len(dbs) > 0 {
		db = dbs[0]
	}

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

	if db != nil {
		s.loadFromDB()
	}
	s.seedData()
	return s
}

func (s *Store) loadFromDB() {
	if s.DB == nil {
		return
	}

	// 1. Hydrate Users
	var users []models.User
	if err := s.DB.Find(&users).Error; err == nil {
		for i := range users {
			s.Users[users[i].ID] = &users[i]
		}
	}

	// 2. Hydrate Restaurants
	var rests []models.Restaurant
	if err := s.DB.Find(&rests).Error; err == nil {
		for i := range rests {
			s.Restaurants[rests[i].ID] = &rests[i]
		}
	}

	// 3. Hydrate Menu Items
	var items []models.MenuItem
	if err := s.DB.Find(&items).Error; err == nil {
		for _, item := range items {
			if item.RestaurantID != "" {
				s.Menus[item.RestaurantID] = append(s.Menus[item.RestaurantID], item)
			}
		}
	}

	// 4. Hydrate Orders
	var orders []models.Order
	if err := s.DB.Find(&orders).Error; err == nil {
		for i := range orders {
			s.Orders[orders[i].ID] = &orders[i]
		}
	}

	// 5. Hydrate Ledger
	var ledger []models.LedgerEntry
	if err := s.DB.Order("created_at asc").Find(&ledger).Error; err == nil {
		s.Ledger = ledger
	}

	// 6. Hydrate Riders
	var riders []models.RiderProfile
	if err := s.DB.Find(&riders).Error; err == nil {
		for i := range riders {
			s.Riders[riders[i].ID] = &riders[i]
		}
	}

	// 7. Hydrate Coins
	var coins []models.CoinProfile
	if err := s.DB.Find(&coins).Error; err == nil {
		for i := range coins {
			s.Coins[coins[i].UserID] = &coins[i]
		}
	}
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

	// Persist to Postgres if database is active
	if s.DB != nil {
		_ = s.DB.Create(&debit)
		_ = s.DB.Create(&credit)
	}

	return nil
}

func (s *Store) SaveOrder(order *models.Order) {
	s.Orders[order.ID] = order
	if s.DB != nil {
		_ = s.DB.Save(order)
	}
}

func (s *Store) SaveRestaurant(r *models.Restaurant) {
	s.Restaurants[r.ID] = r
	if s.DB != nil {
		_ = s.DB.Save(r)
	}
}

func (s *Store) SaveMenuItem(restaurantID string, item *models.MenuItem) {
	item.RestaurantID = restaurantID
	if s.DB != nil {
		_ = s.DB.Save(item)
	}
}

func (s *Store) DeleteMenuItem(restaurantID, itemID string) {
	if s.DB != nil {
		_ = s.DB.Where("id = ? AND restaurant_id = ?", itemID, restaurantID).Delete(&models.MenuItem{})
	}
}

func (s *Store) SaveUser(u *models.User) {
	s.Users[u.ID] = u
	if s.DB != nil {
		_ = s.DB.Save(u)
	}
}

func (s *Store) SaveRider(r *models.RiderProfile) {
	s.Riders[r.ID] = r
	if s.DB != nil {
		_ = s.DB.Save(r)
	}
}

func (s *Store) SaveCoin(c *models.CoinProfile) {
	s.Coins[c.UserID] = c
	if s.DB != nil {
		_ = s.DB.Save(c)
	}
}

func (s *Store) Lock()    { s.mu.Lock() }
func (s *Store) Unlock()  { s.mu.Unlock() }
func (s *Store) RLock()   { s.mu.RLock() }
func (s *Store) RUnlock() { s.mu.RUnlock() }

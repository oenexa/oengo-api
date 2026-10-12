package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"oengo-api/internal/apperrors"
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
		if err := s.loadFromDB(); err != nil {
			log.Printf("[STORE WARNING] Partial error hydrating store from PostgreSQL: %v", err)
		}
	}
	s.seedData()
	return s
}

func (s *Store) loadFromDB() error {
	if s.DB == nil {
		return nil
	}

	// 1. Hydrate Users
	var users []models.User
	if err := s.DB.Find(&users).Error; err != nil {
		return apperrors.WrapDBError("hydrate_users", err)
	}
	for i := range users {
		s.Users[users[i].ID] = &users[i]
	}

	// 2. Hydrate Restaurants
	var rests []models.Restaurant
	if err := s.DB.Find(&rests).Error; err != nil {
		return apperrors.WrapDBError("hydrate_restaurants", err)
	}
	for i := range rests {
		s.Restaurants[rests[i].ID] = &rests[i]
	}

	// 3. Hydrate Menu Items
	var items []models.MenuItem
	if err := s.DB.Find(&items).Error; err != nil {
		return apperrors.WrapDBError("hydrate_menu_items", err)
	}
	for _, item := range items {
		if item.RestaurantID != "" {
			s.Menus[item.RestaurantID] = append(s.Menus[item.RestaurantID], item)
		}
	}

	// 4. Hydrate Orders
	var orders []models.Order
	if err := s.DB.Find(&orders).Error; err != nil {
		return apperrors.WrapDBError("hydrate_orders", err)
	}
	for i := range orders {
		s.Orders[orders[i].ID] = &orders[i]
	}

	// 5. Hydrate Ledger
	var ledger []models.LedgerEntry
	if err := s.DB.Order("created_at asc").Find(&ledger).Error; err != nil {
		return apperrors.WrapDBError("hydrate_ledger", err)
	}
	s.Ledger = ledger

	// 6. Hydrate Riders
	var riders []models.RiderProfile
	if err := s.DB.Find(&riders).Error; err != nil {
		return apperrors.WrapDBError("hydrate_riders", err)
	}
	for i := range riders {
		s.Riders[riders[i].ID] = &riders[i]
	}

	// 7. Hydrate Coins
	var coins []models.CoinProfile
	if err := s.DB.Find(&coins).Error; err != nil {
		return apperrors.WrapDBError("hydrate_coins", err)
	}
	for i := range coins {
		s.Coins[coins[i].UserID] = &coins[i]
	}

	return nil
}

func (s *Store) seedData() {
	// Seed data removed as requested by user. Starting from scratch.
}

func (s *Store) RecordLedgerTx(txID, debitAcc, creditAcc string, amountEUR float64, desc string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if amountEUR <= 0 {
		return apperrors.NewValidation("amountEUR", "ledger transaction amount must be greater than zero")
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

	// Persist atomically to Postgres first if database is active
	if s.DB != nil {
		err := s.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&debit).Error; err != nil {
				return err
			}
			if err := tx.Create(&credit).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			log.Printf("[LEDGER DB ERROR] Failed atomic ledger insert for tx %s: %v", txID, err)
			return apperrors.WrapDBError("RecordLedgerTx", err)
		}
	}

	s.Ledger = append(s.Ledger, debit, credit)
	return nil
}

func (s *Store) SaveOrder(order *models.Order) error {
	if order == nil {
		return apperrors.NewBadRequest("cannot save nil order")
	}
	s.Orders[order.ID] = order
	if s.DB != nil {
		if err := s.DB.Save(order).Error; err != nil {
			log.Printf("[DB ERROR] SaveOrder(%s): %v", order.ID, err)
			return apperrors.WrapDBError("SaveOrder", err)
		}
	}
	return nil
}

func (s *Store) SaveRestaurant(r *models.Restaurant) error {
	if r == nil {
		return apperrors.NewBadRequest("cannot save nil restaurant")
	}
	s.Restaurants[r.ID] = r
	if s.DB != nil {
		if err := s.DB.Save(r).Error; err != nil {
			log.Printf("[DB ERROR] SaveRestaurant(%s): %v", r.ID, err)
			return apperrors.WrapDBError("SaveRestaurant", err)
		}
	}
	return nil
}

func (s *Store) SaveMenuItem(restaurantID string, item *models.MenuItem) error {
	if item == nil {
		return apperrors.NewBadRequest("cannot save nil menu item")
	}
	item.RestaurantID = restaurantID
	if s.DB != nil {
		if err := s.DB.Save(item).Error; err != nil {
			log.Printf("[DB ERROR] SaveMenuItem(%s): %v", item.ID, err)
			return apperrors.WrapDBError("SaveMenuItem", err)
		}
	}
	return nil
}

func (s *Store) DeleteMenuItem(restaurantID, itemID string) error {
	if s.DB != nil {
		if err := s.DB.Where("id = ? AND restaurant_id = ?", itemID, restaurantID).Delete(&models.MenuItem{}).Error; err != nil {
			log.Printf("[DB ERROR] DeleteMenuItem(%s): %v", itemID, err)
			return apperrors.WrapDBError("DeleteMenuItem", err)
		}
	}
	return nil
}

func (s *Store) SaveUser(u *models.User) error {
	if u == nil {
		return apperrors.NewBadRequest("cannot save nil user")
	}
	s.Users[u.ID] = u
	if s.DB != nil {
		if err := s.DB.Save(u).Error; err != nil {
			log.Printf("[DB ERROR] SaveUser(%s): %v", u.ID, err)
			return apperrors.WrapDBError("SaveUser", err)
		}
	}
	return nil
}

func (s *Store) SaveRider(r *models.RiderProfile) error {
	if r == nil {
		return apperrors.NewBadRequest("cannot save nil rider")
	}
	s.Riders[r.ID] = r
	if s.DB != nil {
		if err := s.DB.Save(r).Error; err != nil {
			log.Printf("[DB ERROR] SaveRider(%s): %v", r.ID, err)
			return apperrors.WrapDBError("SaveRider", err)
		}
	}
	return nil
}

func (s *Store) SaveCoin(c *models.CoinProfile) error {
	if c == nil {
		return apperrors.NewBadRequest("cannot save nil coin profile")
	}
	s.Coins[c.UserID] = c
	if s.DB != nil {
		if err := s.DB.Save(c).Error; err != nil {
			log.Printf("[DB ERROR] SaveCoin(%s): %v", c.UserID, err)
			return apperrors.WrapDBError("SaveCoin", err)
		}
	}
	return nil
}

func (s *Store) Lock()    { s.mu.Lock() }
func (s *Store) Unlock()  { s.mu.Unlock() }
func (s *Store) RLock()   { s.mu.RLock() }
func (s *Store) RUnlock() { s.mu.RUnlock() }

package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"oengo-api/internal/config"
	"oengo-api/internal/models"
)

type Store struct {
	mu               sync.RWMutex
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

func NewStore(cfg *config.Config) *Store {
	s := &Store{
		cfg:              cfg,
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
	// Users
	s.Users["user_customer"] = &models.User{
		ID:    "user_customer",
		Name:  "Alice Customer",
		Email: "alice@oengo.delivery",
		Phone: "+39 081 555 0192",
		Role:  models.RoleCustomer,
		DigitalWallet: models.DigitalWallet{
			FiatBalanceEUR:   50.00,
			LockedBalanceEUR: 0.00,
			LoyaltyPoints:    120,
			CashbackRatePct:  5.0,
		},
		CryptoWalletAddress: "0xAlice_Customer_MLDSA65",
	}

	s.Users["user_restaurant"] = &models.User{
		ID:    "user_restaurant",
		Name:  "Napoli Woodfire Pizza",
		Email: "chef@napolipizza.it",
		Phone: "+39 081 777 4421",
		Role:  models.RoleRestaurant,
		DigitalWallet: models.DigitalWallet{
			FiatBalanceEUR: 1250.00,
		},
		CryptoWalletAddress: "0xNapoli_Restaurant_MLDSA65",
	}

	s.Users["user_courier"] = &models.User{
		ID:    "user_courier",
		Name:  "Bob Delivery Rider",
		Email: "bob.rider@oengo.delivery",
		Phone: "+39 081 999 8833",
		Role:  models.RoleCourier,
		DigitalWallet: models.DigitalWallet{
			FiatBalanceEUR: 85.50,
			LoyaltyPoints:  40,
		},
		CryptoWalletAddress: "0xBob_Rider_MLDSA65",
	}

	// Customer Profile
	s.CustomerProfiles["user_customer"] = &models.CustomerProfile{
		ID:     "user_customer",
		Name:   "Alice Customer",
		Email:  "alice@oengo.delivery",
		Phone:  "+39 081 555 0192",
		Avatar: "👩‍💼",
		SavedAddresses: []models.SavedAddress{
			{ID: "addr_1", Title: "Home", Address: "Piazza del Plebiscito 1, Napoli", Lat: 40.8358, Lng: 14.2488, IsDefault: true, Notes: "Ring buzzer #4"},
			{ID: "addr_2", Title: "Work", Address: "Via Toledo 156, Napoli", Lat: 40.8412, Lng: 14.2499, IsDefault: false, Notes: "Reception on 2nd floor"},
		},
		FavoriteRestaurants: []string{"user_restaurant", "rest_burger"},
		FavoriteFoods:       []string{"menu_margherita", "menu_smash_truffle"},
	}

	// Restaurants
	s.Restaurants["user_restaurant"] = &models.Restaurant{
		ID:                  "user_restaurant",
		Name:                "Napoli Woodfire Pizza",
		Tagline:             "Authentic Neapolitan Pizza & Artisan Italian Delicacies",
		Cuisine:             "Italian, Pizza, Artisan",
		Rating:              4.9,
		ReviewCount:         342,
		Address:             "Via Toledo 42, Napoli / Historic District",
		Icon:                "🍕",
		IsOpen:              true,
		PrepEtaMinutes:      15,
		DeliveryFeeEUR:      2.50,
		MinOrderEUR:         15.00,
		DeliveryRadiusKm:    5.5,
		TaxNumber:           "IT98765432101",
		VATNumber:           "IT98765432101",
		CryptoWalletAddress: "0xNapoli_Restaurant_MLDSA65",
		FiatBalanceEUR:      1250.00,
	}

	s.Restaurants["rest_burger"] = &models.Restaurant{
		ID:                  "rest_burger",
		Name:                "Smash & Co. Gourmet Burgers",
		Tagline:             "Double Smashed Dry-Aged Angus & Loaded Brioche",
		Cuisine:             "American, Burgers, Craft",
		Rating:              4.8,
		ReviewCount:         218,
		Address:             "Corso Umberto I 118, Napoli",
		Icon:                "🍔",
		IsOpen:              true,
		PrepEtaMinutes:      20,
		DeliveryFeeEUR:      2.99,
		MinOrderEUR:         12.00,
		DeliveryRadiusKm:    6.0,
		TaxNumber:           "IT12345678902",
		VATNumber:           "IT12345678902",
		CryptoWalletAddress: "0xSmashBurger_MLDSA65",
		FiatBalanceEUR:      890.00,
	}

	s.Restaurants["rest_sushi"] = &models.Restaurant{
		ID:                  "rest_sushi",
		Name:                "Tokyo Bloom Omakase & Bento",
		Tagline:             "Sustainably Sourced Sashimi, Nigiri & Crispy Gyoza",
		Cuisine:             "Japanese, Sushi, Asian",
		Rating:              4.9,
		ReviewCount:         451,
		Address:             "Via Chiaia 84, Napoli",
		Icon:                "🍣",
		IsOpen:              true,
		PrepEtaMinutes:      25,
		DeliveryFeeEUR:      3.50,
		MinOrderEUR:         20.00,
		DeliveryRadiusKm:    7.0,
		TaxNumber:           "IT55443322119",
		VATNumber:           "IT55443322119",
		CryptoWalletAddress: "0xTokyoBloom_MLDSA65",
		FiatBalanceEUR:      2140.00,
	}

	s.Restaurants["rest_green"] = &models.Restaurant{
		ID:                  "rest_green",
		Name:                "Verde Organic Bowls & Cold Press",
		Tagline:             "Superfood Macro Bowls, Wild Greens & Fresh Smoothies",
		Cuisine:             "Healthy, Vegan, Organic",
		Rating:              4.7,
		ReviewCount:         129,
		Address:             "Piazza dei Martiri 16, Napoli",
		Icon:                "🥗",
		IsOpen:              true,
		PrepEtaMinutes:      12,
		DeliveryFeeEUR:      1.99,
		MinOrderEUR:         10.00,
		DeliveryRadiusKm:    4.5,
		TaxNumber:           "IT77665544338",
		VATNumber:           "IT77665544338",
		CryptoWalletAddress: "0xVerdeOrganic_MLDSA65",
		FiatBalanceEUR:      620.00,
	}

	defaultCustomizations := []models.CustomizationGroup{
		{
			Name:    "Choose Bread / Crust",
			IsMulti: false,
			Options: []models.CustomizationOption{
				{Name: "Traditional Artisan", ExtraEUR: 0.00},
				{Name: "Whole Grain / Rustic", ExtraEUR: 1.00},
				{Name: "Gluten-Free Certified", ExtraEUR: 2.50},
			},
		},
		{
			Name:    "Portion Size",
			IsMulti: false,
			Options: []models.CustomizationOption{
				{Name: "Standard Regular", ExtraEUR: 0.00},
				{Name: "Large (+30% portion)", ExtraEUR: 3.50},
			},
		},
		{
			Name:    "Extra Toppings & Sauces",
			IsMulti: true,
			Options: []models.CustomizationOption{
				{Name: "Extra Buffalo Mozzarella", ExtraEUR: 2.00},
				{Name: "Crispy Pancetta / Bacon", ExtraEUR: 2.20},
				{Name: "Truffle Glaze Infusion", ExtraEUR: 1.80},
			},
		},
	}

	// Menus
	s.Menus["user_restaurant"] = []models.MenuItem{
		{
			ID:                "menu_margherita",
			Category:          "Pizza & Mains",
			Name:              "Artisanal Margherita Pizza",
			Description:       "San Marzano D.O.P. tomatoes, fresh buffalo mozzarella, fragrant basil, extra virgin olive oil.",
			PriceEUR:          16.50,
			PriceOEN:          "1.21",
			InStock:           true,
			StockQuantity:     45,
			LowStockThreshold: 5,
			PrepMinutes:       12,
			Badge:             "Bestseller",
			Calories:          780,
			Ingredients:       "Flour, San Marzano Tomatoes, Buffalo Mozzarella, Basil, Olive Oil",
			Customizations:    defaultCustomizations,
		},
		{
			ID:                "menu_diavola",
			Category:          "Pizza & Mains",
			Name:              "Spicy Diavola Pizza",
			Description:       "Spianata Calabrese spicy salami, smoked provolone, chili flakes, organic tomato reduction.",
			PriceEUR:          18.00,
			PriceOEN:          "1.32",
			InStock:           true,
			StockQuantity:     28,
			LowStockThreshold: 5,
			PrepMinutes:       14,
			Badge:             "Spicy",
			Calories:          890,
			Ingredients:       "Flour, Spicy Salami, Smoked Provolone, Tomato Sauce, Chili Oil",
			Customizations:    defaultCustomizations,
		},
		{
			ID:                "menu_arancini",
			Category:          "Starters",
			Name:              "Truffle & Porcini Arancini",
			Description:       "Crispy golden saffron risotto balls stuffed with black truffle cream and melted fontina cheese.",
			PriceEUR:          12.00,
			PriceOEN:          "0.88",
			InStock:           true,
			StockQuantity:     18,
			LowStockThreshold: 4,
			PrepMinutes:       8,
			Badge:             "Vegetarian",
			Calories:          450,
			Ingredients:       "Carnaroli Rice, Porcini, Black Truffle, Fontina, Saffron",
			Customizations:    nil,
		},
		{
			ID:                "menu_tiramisu",
			Category:          "Desserts",
			Name:              "Classic Espresso Tiramisù",
			Description:       "Layered savoiardi soaked in single-origin espresso and Marsala, mascarpone cream, dark cocoa.",
			PriceEUR:          8.50,
			PriceOEN:          "0.62",
			InStock:           true,
			StockQuantity:     22,
			LowStockThreshold: 5,
			PrepMinutes:       5,
			Badge:             "Homemade",
			Calories:          410,
			Ingredients:       "Savoiardi, Espresso, Mascarpone, Free-range Eggs, Cocoa Powder",
			Customizations:    nil,
		},
	}

	s.Menus["rest_burger"] = []models.MenuItem{
		{
			ID:                "menu_smash_truffle",
			Category:          "Burgers",
			Name:              "Double Truffle Smash Burger",
			Description:       "Two 100g dry-aged Angus patties, double American cheese, black truffle aioli, grilled onions, brioche.",
			PriceEUR:          15.50,
			PriceOEN:          "1.14",
			InStock:           true,
			StockQuantity:     30,
			LowStockThreshold: 5,
			PrepMinutes:       12,
			Badge:             "Bestseller",
			Calories:          880,
			Ingredients:       "Black Angus Beef, Brioche Bun, Truffle Aioli, American Cheddar, Caramelized Onion",
			Customizations:    defaultCustomizations,
		},
		{
			ID:                "menu_parm_fries",
			Category:          "Sides",
			Name:              "Parmesan & Rosemary Fries",
			Description:       "Triple-cooked rustic skin-on potatoes dusted with 24-month Parmigiano Reggiano and fresh rosemary.",
			PriceEUR:          6.00,
			PriceOEN:          "0.44",
			InStock:           true,
			StockQuantity:     50,
			LowStockThreshold: 10,
			PrepMinutes:       6,
			Badge:             "Crispy",
			Calories:          380,
			Ingredients:       "Skin-on Potatoes, Parmigiano Reggiano, Rosemary, Sea Salt",
			Customizations:    nil,
		},
	}

	// Seed Order
	seedOrderID := "ord_seed_101"
	s.Orders[seedOrderID] = &models.Order{
		ID:                seedOrderID,
		BuyerID:           "user_customer",
		BuyerName:         "Alice Customer",
		BuyerAddress:      "0xAlice_Customer_MLDSA65",
		RestaurantID:      "user_restaurant",
		RestaurantAddress: "0xNapoli_Restaurant_MLDSA65",
		CourierID:         "user_courier",
		CourierAddress:    "0xBob_Rider_MLDSA65",
		Items: []models.OrderItem{
			{Name: "Artisanal Margherita Pizza", Qty: 1, Price: 16.50},
			{Name: "Truffle & Porcini Arancini", Qty: 1, Price: 12.00},
		},
		Amount:            28.50,
		DeliveryFee:       3.50,
		Tip:               2.00,
		Total:             34.00,
		CommissionPct:     5.0,
		PaymentMethod:     "CREDIT_CARD",
		CardPayment:       &models.CardPaymentInfo{Brand: "Visa", Last4: "4242", TransactionID: "txn_seed_4242"},
		PaymentStatus:     "PAID",
		Status:            models.StatusAwaitingRestaurant,
		PickupBarcode:     "PKG-DEMO01",
		PickupBarcodeHash: Sha256Hash("PKG-DEMO01"),
		DeliveryPIN:       "4821",
		DeliveryPINHash:   Sha256Hash("4821"),
		CreatedAt:         time.Now().Add(-4 * time.Minute).Format(time.RFC3339),
		EscrowLocked:      true,
	}

	// Double-entry initial entries
	s.Ledger = append(s.Ledger,
		models.LedgerEntry{
			ID:          "ledg_init_001",
			TxID:        "tx_seed_dep",
			AccountCode: "1001", // Cash Clearing
			EntryType:   models.LedgerDebit,
			AmountEUR:   50.00,
			Description: "Initial customer deposit clearing",
			CreatedAt:   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		},
		models.LedgerEntry{
			ID:          "ledg_init_002",
			TxID:        "tx_seed_dep",
			AccountCode: "1002", // Customer Available Wallet
			EntryType:   models.LedgerCredit,
			AmountEUR:   50.00,
			Description: "Customer digital wallet credited",
			CreatedAt:   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		},
	)

	// Coins
	s.Coins["user_customer"] = &models.CoinProfile{
		UserID:      "user_customer",
		CoinBalance: 1250,
		TotalEarned: 1500,
		TotalSpent:  250,
		Transactions: []models.CoinTx{
			{ID: "coin_tx_1", Type: "EARNED", Coins: 100, Reason: "SIGNUP_BONUS", CreatedAt: time.Now().Add(-7 * 24 * time.Hour)},
			{ID: "coin_tx_2", Type: "EARNED", Coins: 250, Reason: "REFERRAL_REWARD", CreatedAt: time.Now().Add(-3 * 24 * time.Hour)},
			{ID: "coin_tx_3", Type: "EARNED", Coins: 900, Reason: "ORDER_CASHBACK_5PCT", CreatedAt: time.Now().Add(-1 * 24 * time.Hour)},
		},
	}

	// Riders
	s.Riders["user_courier"] = &models.RiderProfile{
		ID:               "user_courier",
		Name:             "Bob Delivery Rider",
		Phone:            "+39 081 999 8833",
		IsOnline:         true,
		VehicleType:      "Electric Bicycle",
		VehiclePlate:     "E-BIKE-NA104",
		KYCStatus:        "VERIFIED",
		Rating:           4.98,
		TotalDeliveries:  184,
		TodayEarningsEUR: 42.50,
		CurrentLocation: models.Location{
			Lat:       40.8385,
			Lng:       14.2505,
			Heading:   65.0,
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	}

	// Coupons
	disc5 := 5.0
	s.Coupons["WELCOME5"] = &models.Coupon{
		Code:        "WELCOME5",
		DiscountEUR: &disc5,
		MinOrderEUR: 15.00,
		MaxUses:     1000,
		TimesUsed:   14,
		IsActive:    true,
	}

	discPct10 := 10.0
	s.Coupons["TASTY10"] = &models.Coupon{
		Code:        "TASTY10",
		DiscountPct: &discPct10,
		MinOrderEUR: 20.00,
		MaxUses:     500,
		TimesUsed:   8,
		IsActive:    true,
	}
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

func (s *Store) Lock()   { s.mu.Lock() }
func (s *Store) Unlock() { s.mu.Unlock() }
func (s *Store) RLock()  { s.mu.RLock() }
func (s *Store) RUnlock() { s.mu.RUnlock() }

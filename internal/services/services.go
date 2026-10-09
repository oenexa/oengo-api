package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"

	"oengo-api/internal/config"
	"oengo-api/internal/models"
	"oengo-api/internal/store"
)

type Services struct {
	Store    *store.Store
	Cfg      *config.Config
	Notifier *NotificationService
}

func NewServices(s *store.Store, cfg *config.Config) *Services {
	// Initialize the Notification Service (with mock keys for now)
	notifier := NewNotificationService("SG.MockKey123", "ACMockTwilioSID", "MockTwilioAuth")
	return &Services{Store: s, Cfg: cfg, Notifier: notifier}
}

// ── RESTAURANT SERVICE ──────────────────────────────────────────────────────

func (svc *Services) ListRestaurants(cuisine, search string) []*models.Restaurant {
	svc.Store.RLock()
	defer svc.Store.RUnlock()

	var result []*models.Restaurant
	cLower := strings.ToLower(cuisine)
	sLower := strings.ToLower(search)

	for _, r := range svc.Store.Restaurants {
		if cuisine != "" && cuisine != "ALL" && !strings.Contains(strings.ToLower(r.Cuisine), cLower) {
			continue
		}
		if search != "" {
			nameMatch := strings.Contains(strings.ToLower(r.Name), sLower)
			cuiMatch := strings.Contains(strings.ToLower(r.Cuisine), sLower)
			tagMatch := strings.Contains(strings.ToLower(r.Tagline), sLower)
			if !nameMatch && !cuiMatch && !tagMatch {
				continue
			}
		}
		result = append(result, r)
	}
	return result
}

func (svc *Services) GetRestaurantProfile(id string) (*models.Restaurant, error) {
	svc.Store.RLock()
	defer svc.Store.RUnlock()

	r, exists := svc.Store.Restaurants[id]
	if !exists {
		return nil, fmt.Errorf("restaurant not found")
	}

	// Compute live revenue stats
	var totalCount, activeCount, deliveredCount int
	var grossDelivered, activePipeline float64

	for _, o := range svc.Store.Orders {
		if o.RestaurantID == id {
			totalCount++
			if o.Status == models.StatusDelivered {
				deliveredCount++
				if o.RestaurantPayout > 0 {
					grossDelivered += o.RestaurantPayout
				} else {
					grossDelivered += o.Amount * 0.95
				}
			} else if o.Status == models.StatusAwaitingRestaurant || o.Status == models.StatusPreparing ||
				o.Status == models.StatusReadyForPickup || o.Status == models.StatusInTransit {
				activeCount++
				activePipeline += o.Amount * ((100.0 - o.CommissionPct) / 100.0)
			}
		}
	}

	grossToday := math.Round((grossDelivered+activePipeline)*100) / 100
	retainedPct := 100.0 - svc.Store.CommissionPct
	lostRevenue := math.Round((grossToday*(svc.Cfg.LegacyAggregatorFee-svc.Store.CommissionPct)/100.0)*100) / 100

	rCopy := *r
	rCopy.Stats = &models.RestaurantStats{
		TotalOrdersCount:      totalCount,
		ActiveOrdersCount:     activeCount,
		DeliveredOrdersCount:  deliveredCount,
		GrossRevenueTodayEUR:  grossToday,
		FiatBalanceEUR:        r.FiatBalanceEUR,
		CommissionRetainedPct: retainedPct,
		PlatformCommissionPct: svc.Store.CommissionPct,
		LegacyLostRevenueEUR:  lostRevenue,
		AvgPrepTimeMinutes:    r.PrepEtaMinutes,
	}

	return &rCopy, nil
}

func (svc *Services) UpdateRestaurant(id string, isOpen *bool, prepEta *int, name, tagline *string) (*models.Restaurant, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	r, exists := svc.Store.Restaurants[id]
	if !exists {
		return nil, fmt.Errorf("restaurant not found")
	}

	if isOpen != nil {
		r.IsOpen = *isOpen
	}
	if prepEta != nil {
		r.PrepEtaMinutes = *prepEta
	}
	if name != nil && *name != "" {
		r.Name = *name
	}
	if tagline != nil && *tagline != "" {
		r.Tagline = *tagline
	}

	return r, nil
}

func (svc *Services) GetMenu(restaurantID string) []models.MenuItem {
	svc.Store.RLock()
	defer svc.Store.RUnlock()

	return svc.Store.Menus[restaurantID]
}

func (svc *Services) AddDish(restaurantID string, item models.MenuItem) (*models.MenuItem, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	if item.Name == "" || item.PriceEUR <= 0 {
		return nil, fmt.Errorf("dish name and positive price are required")
	}

	item.ID = fmt.Sprintf("menu_%d", time.Now().UnixMilli())
	item.PriceOEN = fmt.Sprintf("%.2f", item.PriceEUR/svc.Cfg.OenEurExchangeRate)
	item.InStock = true
	if item.StockQuantity == 0 {
		item.StockQuantity = 20
	}
	if item.LowStockThreshold == 0 {
		item.LowStockThreshold = 5
	}
	if item.PrepMinutes == 0 {
		item.PrepMinutes = 15
	}

	svc.Store.Menus[restaurantID] = append(svc.Store.Menus[restaurantID], item)
	return &item, nil
}

func (svc *Services) UpdateDish(restaurantID, itemID string, inStock *bool, price *float64, stockQty *int) (*models.MenuItem, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	menu := svc.Store.Menus[restaurantID]
	for i := range menu {
		if menu[i].ID == itemID {
			if inStock != nil {
				menu[i].InStock = *inStock
			}
			if price != nil && *price > 0 {
				menu[i].PriceEUR = *price
				menu[i].PriceOEN = fmt.Sprintf("%.2f", *price/svc.Cfg.OenEurExchangeRate)
			}
			if stockQty != nil {
				menu[i].StockQuantity = *stockQty
				if *stockQty <= 0 {
					menu[i].InStock = false
				}
			}
			return &menu[i], nil
		}
	}
	return nil, fmt.Errorf("menu item not found")
}

func (svc *Services) DeleteDish(restaurantID, itemID string) error {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	menu := svc.Store.Menus[restaurantID]
	for i, item := range menu {
		if item.ID == itemID {
			svc.Store.Menus[restaurantID] = append(menu[:i], menu[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("menu item not found")
}

// ── ORDER SERVICE ───────────────────────────────────────────────────────────

type CreateOrderRequest struct {
	BuyerID       string                  `json:"buyerId"`
	RestaurantID  string                  `json:"restaurantId"`
	Items         []models.OrderItem      `json:"items"`
	Amount        float64                 `json:"amount"`
	DeliveryFee   float64                 `json:"deliveryFee"`
	Tip           float64                 `json:"tip"`
	PaymentMethod string                  `json:"paymentMethod"`
	CardPayment   *models.CardPaymentInfo `json:"cardPayment"`
	CommissionPct *float64                `json:"commissionPct"`
}

func (svc *Services) CreateOrder(req CreateOrderRequest) (*models.Order, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	if req.BuyerID == "" {
		req.BuyerID = "user_customer"
	}
	if req.RestaurantID == "" {
		req.RestaurantID = "user_restaurant"
	}

	buyer, exists := svc.Store.Users[req.BuyerID]
	if !exists {
		return nil, fmt.Errorf("invalid buyer ID")
	}
	rest, exists := svc.Store.Restaurants[req.RestaurantID]
	if !exists {
		return nil, fmt.Errorf("invalid restaurant ID")
	}

	commPct := svc.Store.CommissionPct
	if req.CommissionPct != nil && *req.CommissionPct >= svc.Cfg.MinCommissionPct && *req.CommissionPct <= svc.Cfg.MaxCommissionPct {
		commPct = *req.CommissionPct
	}

	orderID := fmt.Sprintf("ord_%d", time.Now().UnixMilli())
	barcodeSuffix := strings.ToUpper(orderID[len(orderID)-6:])
	pickupBarcode := fmt.Sprintf("PKG-%s", barcodeSuffix)
	deliveryPIN := fmt.Sprintf("%04d", time.Now().UnixNano()%9000+1000)

	total := req.Amount + req.DeliveryFee + req.Tip

	order := &models.Order{
		ID:                orderID,
		BuyerID:           req.BuyerID,
		BuyerName:         buyer.Name,
		BuyerAddress:      buyer.CryptoWalletAddress,
		RestaurantID:      req.RestaurantID,
		RestaurantAddress: rest.CryptoWalletAddress,
		Items:             req.Items,
		Amount:            req.Amount,
		DeliveryFee:       req.DeliveryFee,
		Tip:               req.Tip,
		Total:             total,
		CommissionPct:     commPct,
		PaymentMethod:     req.PaymentMethod,
		CardPayment:       req.CardPayment,
		PaymentStatus:     "PAID",
		Status:            models.StatusAwaitingRestaurant,
		PickupBarcode:     pickupBarcode,
		PickupBarcodeHash: store.Sha256Hash(pickupBarcode),
		DeliveryPIN:       deliveryPIN,
		DeliveryPINHash:   store.Sha256Hash(deliveryPIN),
		CreatedAt:         time.Now().Format(time.RFC3339),
		EscrowLocked:      true,
	}

	svc.Store.Orders[orderID] = order

	// Double-entry: Debit 1001 (Cash Clearing) / Credit 1003 (Locked Escrow)
	go svc.Store.RecordLedgerTx(orderID, "1001", "1003", total, fmt.Sprintf("Order %s escrow lock", orderID))

	return order, nil
}

func (svc *Services) GetOrder(id string) (*models.Order, error) {
	svc.Store.RLock()
	defer svc.Store.RUnlock()

	o, exists := svc.Store.Orders[id]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}
	return o, nil
}

func (svc *Services) ListOrders(restaurantID, buyerID, status string) []*models.Order {
	svc.Store.RLock()
	defer svc.Store.RUnlock()

	var result []*models.Order
	for _, o := range svc.Store.Orders {
		if restaurantID != "" && o.RestaurantID != restaurantID {
			continue
		}
		if buyerID != "" && o.BuyerID != buyerID {
			continue
		}
		if status != "" && string(o.Status) != status {
			continue
		}
		result = append(result, o)
	}
	return result
}

func (svc *Services) UpdateOrderStatus(id string, status models.OrderStatus) (*models.Order, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	o, exists := svc.Store.Orders[id]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}

	o.Status = status

	// Notify Customer via Email/SMS
	if svc.Notifier != nil {
		user, hasUser := svc.Store.Users[o.BuyerID]
		if hasUser {
			svc.Notifier.NotifyOrderStatusChange(user.Email, user.Phone, o.ID, string(o.Status))
		}
	}

	return o, nil
}

func (svc *Services) AssignCourier(orderID, courierID string) (*models.Order, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	o, exists := svc.Store.Orders[orderID]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}

	courier, exists := svc.Store.Users[courierID]
	if !exists {
		return nil, fmt.Errorf("courier not found")
	}

	o.CourierID = courierID
	o.CourierAddress = courier.CryptoWalletAddress
	return o, nil
}

func (svc *Services) ConfirmPickup(orderID, barcode string) (*models.Order, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	o, exists := svc.Store.Orders[orderID]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}

	if o.PickupBarcode != barcode && o.PickupBarcodeHash != store.Sha256Hash(barcode) {
		return nil, fmt.Errorf("invalid pickup barcode")
	}

	o.Status = models.StatusInTransit
	return o, nil
}

func (svc *Services) ConfirmDelivery(orderID, pinOrBarcode string) (*models.Order, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	o, exists := svc.Store.Orders[orderID]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}

	// Validate PIN or Barcode
	pinMatch := o.DeliveryPIN == pinOrBarcode || o.DeliveryPINHash == store.Sha256Hash(pinOrBarcode)
	barcodeMatch := o.PickupBarcode == pinOrBarcode || o.PickupBarcodeHash == store.Sha256Hash(pinOrBarcode)
	if !pinMatch && !barcodeMatch {
		return nil, fmt.Errorf("invalid delivery PIN or barcode")
	}

	// Settlement calculations:
	// Platform commission = Amount * (CommissionPct / 100)
	// Restaurant payout = Amount - Platform commission
	// Courier payout = Platform commission + DeliveryFee + Tip
	commAmount := math.Round(o.Amount*(o.CommissionPct/100.0)*100) / 100
	restPayout := math.Round((o.Amount-commAmount)*100) / 100
	courierPayout := math.Round((commAmount+o.DeliveryFee+o.Tip)*100) / 100

	o.Status = models.StatusDelivered
	o.EscrowLocked = false
	o.DeliveredAt = time.Now().Format(time.RFC3339)
	o.RestaurantPayout = restPayout
	o.CourierPayout = courierPayout

	// Update restaurant and rider fiat balances
	if rest, ok := svc.Store.Restaurants[o.RestaurantID]; ok {
		rest.FiatBalanceEUR += restPayout
	}
	if rider, ok := svc.Store.Riders[o.CourierID]; ok {
		rider.TodayEarningsEUR += courierPayout
		rider.TotalDeliveries++
	}

	// Double-entry multi-split ledger:
	// Debit 1003 (Locked Escrow): Total
	// Credit 2001 (Restaurant Payable): restPayout
	// Credit 2002 (Courier Payable): courierPayout
	go svc.Store.RecordLedgerTx(orderID, "1003", "2001", restPayout, fmt.Sprintf("Order %s delivery restaurant payout", orderID))
	go svc.Store.RecordLedgerTx(orderID, "1003", "2002", courierPayout, fmt.Sprintf("Order %s delivery courier payout", orderID))

	// Mint 5% loyalty coins for customer
	cashbackCoins := int64(math.Round(o.Amount * 0.05 * 100))
	if profile, ok := svc.Store.Coins[o.BuyerID]; ok {
		profile.CoinBalance += cashbackCoins
		profile.TotalEarned += cashbackCoins
		profile.Transactions = append(profile.Transactions, models.CoinTx{
			ID:        fmt.Sprintf("ctx_%d", time.Now().UnixNano()),
			Type:      "EARNED",
			Coins:     cashbackCoins,
			Reason:    "ORDER_CASHBACK_5PCT",
			OrderID:   orderID,
			CreatedAt: time.Now(),
		})
	}

	return o, nil
}

func (svc *Services) DeclineOrder(orderID, reason string) (*models.Order, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	o, exists := svc.Store.Orders[orderID]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}

	o.Status = models.StatusRefunded
	o.EscrowLocked = false

	// Refund buyer wallet if applicable
	if buyer, ok := svc.Store.Users[o.BuyerID]; ok {
		buyer.DigitalWallet.FiatBalanceEUR += o.Total
	}

	// Double-entry: Debit 1003 (Locked Escrow) / Credit 1002 (Customer Available Wallet)
	go svc.Store.RecordLedgerTx(orderID, "1003", "1002", o.Total, fmt.Sprintf("Order %s refund: %s", orderID, reason))

	return o, nil
}

// ── WALLET & LEDGER SERVICE ─────────────────────────────────────────────────

func (svc *Services) GetWallet(userID string) (*models.User, error) {
	svc.Store.RLock()
	defer svc.Store.RUnlock()

	u, exists := svc.Store.Users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}

func (svc *Services) DepositWallet(userID string, amount float64) (float64, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	u, exists := svc.Store.Users[userID]
	if !exists {
		return 0, fmt.Errorf("user not found")
	}
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be greater than zero")
	}

	u.DigitalWallet.FiatBalanceEUR += amount

	txID := fmt.Sprintf("dep_%d", time.Now().UnixMilli())
	go svc.Store.RecordLedgerTx(txID, "1001", "1002", amount, fmt.Sprintf("User %s wallet deposit", userID))

	return u.DigitalWallet.FiatBalanceEUR, nil
}

func (svc *Services) WithdrawWallet(userID string, amount float64, iban string) (float64, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	u, exists := svc.Store.Users[userID]
	if !exists {
		return 0, fmt.Errorf("user not found")
	}
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be greater than zero")
	}
	if u.DigitalWallet.FiatBalanceEUR < amount {
		return 0, fmt.Errorf("insufficient wallet balance")
	}

	u.DigitalWallet.FiatBalanceEUR -= amount

	txID := fmt.Sprintf("wth_%d", time.Now().UnixMilli())
	go svc.Store.RecordLedgerTx(txID, "1002", "1001", amount, fmt.Sprintf("User %s withdrawal to %s", userID, iban))

	return u.DigitalWallet.FiatBalanceEUR, nil
}

func (svc *Services) GetLedgerEntries(limit int) []models.LedgerEntry {
	svc.Store.RLock()
	defer svc.Store.RUnlock()

	if limit <= 0 || limit > len(svc.Store.Ledger) {
		limit = len(svc.Store.Ledger)
	}

	// Return most recent first
	res := make([]models.LedgerEntry, limit)
	total := len(svc.Store.Ledger)
	for i := 0; i < limit; i++ {
		res[i] = svc.Store.Ledger[total-1-i]
	}
	return res
}

// ── COIN SERVICE ────────────────────────────────────────────────────────────

func (svc *Services) GetCoinProfile(userID string) *models.CoinProfile {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	profile, exists := svc.Store.Coins[userID]
	if !exists {
		profile = &models.CoinProfile{
			UserID:      userID,
			CoinBalance: 100, // Welcome signup bonus
			TotalEarned: 100,
			TotalSpent:  0,
			Transactions: []models.CoinTx{
				{ID: fmt.Sprintf("ctx_%d", time.Now().UnixNano()), Type: "EARNED", Coins: 100, Reason: "SIGNUP_BONUS", CreatedAt: time.Now()},
			},
		}
		svc.Store.Coins[userID] = profile
	}
	return profile
}

func (svc *Services) RedeemCoins(userID string, coins int64, subtotal float64) (float64, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	profile, exists := svc.Store.Coins[userID]
	if !exists || profile.CoinBalance < coins {
		return 0, fmt.Errorf("insufficient OENGO coins")
	}

	// 100 coins = €1.00
	discountEUR := float64(coins) / 100.0
	maxDiscount := subtotal * 0.20 // Max 20% order subtotal
	if subtotal > 0 && discountEUR > maxDiscount {
		return 0, fmt.Errorf("exceeds max coin discount of 20%% (€%.2f)", maxDiscount)
	}

	profile.CoinBalance -= coins
	profile.TotalSpent += coins
	profile.Transactions = append(profile.Transactions, models.CoinTx{
		ID:        fmt.Sprintf("ctx_%d", time.Now().UnixNano()),
		Type:      "REDEEMED",
		Coins:     coins,
		Reason:    "ORDER_DISCOUNT",
		CreatedAt: time.Now(),
	})

	return discountEUR, nil
}

// ── PAYMENT SERVICE ─────────────────────────────────────────────────────────

func (svc *Services) CreateCardIntent(amount float64, currency string) (map[string]interface{}, error) {
	b := make([]byte, 4)
	rand.Read(b)
	intentID := fmt.Sprintf("pi_%d_%s", time.Now().UnixMilli(), hex.EncodeToString(b))

	return map[string]interface{}{
		"paymentIntentId": intentID,
		"clientSecret":    fmt.Sprintf("sec_%x", b),
		"amount":          amount,
		"currency":        currency,
		"status":          "REQUIRES_PAYMENT_METHOD",
	}, nil
}

func (svc *Services) ConfirmCardPayment(intentID, cardNumber, cardHolder string, saveCard bool, userID string) (*models.CardPaymentInfo, error) {
	cleanNum := strings.ReplaceAll(cardNumber, " ", "")
	if len(cleanNum) < 13 {
		return nil, fmt.Errorf("invalid card number length")
	}

	brand := "Visa"
	if strings.HasPrefix(cleanNum, "5") {
		brand = "Mastercard"
	} else if strings.HasPrefix(cleanNum, "3") {
		brand = "American Express"
	}

	last4 := cleanNum[len(cleanNum)-4:]
	txID := fmt.Sprintf("txn_%d", time.Now().UnixMilli())

	cardInfo := &models.CardPaymentInfo{
		Brand:         brand,
		Last4:         last4,
		TransactionID: txID,
	}

	if saveCard {
		svc.Store.Lock()
		svc.Store.SavedCards[userID] = append(svc.Store.SavedCards[userID], *cardInfo)
		svc.Store.Unlock()
	}

	return cardInfo, nil
}

func (svc *Services) InstantPay(userID, method string, amount float64) (map[string]interface{}, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	user, exists := svc.Store.Users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	if method == "WALLET" && user.DigitalWallet.FiatBalanceEUR < amount {
		return nil, fmt.Errorf("insufficient wallet balance")
	}

	txID := fmt.Sprintf("inst_%d", time.Now().UnixMilli())
	if method == "WALLET" {
		user.DigitalWallet.FiatBalanceEUR -= amount
		go svc.Store.RecordLedgerTx(txID, "1002", "1001", amount, "Instant 1-Click Pay debit")
	}

	return map[string]interface{}{
		"success":       true,
		"transactionId": txID,
		"method":        method,
		"amountEUR":     amount,
		"status":        "SUCCEEDED",
		"timestamp":     time.Now().Format(time.RFC3339),
	}, nil
}

// ── RIDER SERVICE ───────────────────────────────────────────────────────────

func (svc *Services) GetRiderProfile(riderID string) (*models.RiderProfile, error) {
	svc.Store.RLock()
	defer svc.Store.RUnlock()

	r, exists := svc.Store.Riders[riderID]
	if !exists {
		return nil, fmt.Errorf("rider not found")
	}
	return r, nil
}

func (svc *Services) ToggleRiderStatus(riderID string, isOnline bool) (*models.RiderProfile, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	r, exists := svc.Store.Riders[riderID]
	if !exists {
		return nil, fmt.Errorf("rider not found")
	}
	r.IsOnline = isOnline
	return r, nil
}

func (svc *Services) UpdateRiderLocation(riderID string, lat, lng, heading float64) (*models.Location, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	r, exists := svc.Store.Riders[riderID]
	if !exists {
		return nil, fmt.Errorf("rider not found")
	}
	r.CurrentLocation = models.Location{
		Lat:       lat,
		Lng:       lng,
		Heading:   heading,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
	return &r.CurrentLocation, nil
}

func (svc *Services) GetRiderJobs(riderID string) []*models.Order {
	svc.Store.RLock()
	defer svc.Store.RUnlock()

	var jobs []*models.Order
	for _, o := range svc.Store.Orders {
		if o.Status == models.StatusReadyForPickup || o.Status == models.StatusPreparing {
			jobs = append(jobs, o)
		}
	}
	return jobs
}

// ── CUSTOMER SERVICE ────────────────────────────────────────────────────────

func (svc *Services) GetCustomerProfile(userID string) (*models.CustomerProfile, error) {
	svc.Store.RLock()
	defer svc.Store.RUnlock()

	p, exists := svc.Store.CustomerProfiles[userID]
	if !exists {
		return nil, fmt.Errorf("customer profile not found")
	}
	return p, nil
}

func (svc *Services) AddSavedAddress(userID string, addr models.SavedAddress) (*models.CustomerProfile, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	p, exists := svc.Store.CustomerProfiles[userID]
	if !exists {
		return nil, fmt.Errorf("customer profile not found")
	}

	addr.ID = fmt.Sprintf("addr_%d", time.Now().UnixMilli())
	p.SavedAddresses = append(p.SavedAddresses, addr)
	return p, nil
}

// ── ADMIN SERVICE ───────────────────────────────────────────────────────────

type AdminDashboardStats struct {
	TotalOrdersCount      int     `json:"totalOrdersCount"`
	ActiveOrdersCount     int     `json:"activeOrdersCount"`
	TotalGMVEUR           float64 `json:"totalGmvEUR"`
	PlatformCommissionEUR float64 `json:"platformCommissionEUR"`
	MerchantRetentionEUR  float64 `json:"merchantRetentionEUR"`
	ActiveRidersCount     int     `json:"activeRidersCount"`
	DefaultCommissionPct  float64 `json:"defaultCommissionPct"`
}

func (svc *Services) GetAdminDashboard() *AdminDashboardStats {
	svc.Store.RLock()
	defer svc.Store.RUnlock()

	var totalOrders, activeOrders int
	var totalGMV, platformComm, merchantRet float64

	for _, o := range svc.Store.Orders {
		totalOrders++
		totalGMV += o.Total
		if o.Status != models.StatusDelivered && o.Status != models.StatusCancelled && o.Status != models.StatusRefunded {
			activeOrders++
		}
		comm := o.Amount * (o.CommissionPct / 100.0)
		platformComm += comm
		merchantRet += (o.Amount - comm)
	}

	activeRiders := 0
	for _, r := range svc.Store.Riders {
		if r.IsOnline {
			activeRiders++
		}
	}

	return &AdminDashboardStats{
		TotalOrdersCount:      totalOrders,
		ActiveOrdersCount:     activeOrders,
		TotalGMVEUR:           math.Round(totalGMV*100) / 100,
		PlatformCommissionEUR: math.Round(platformComm*100) / 100,
		MerchantRetentionEUR:  math.Round(merchantRet*100) / 100,
		ActiveRidersCount:     activeRiders,
		DefaultCommissionPct:  svc.Store.CommissionPct,
	}
}

func (svc *Services) SetCommissionRate(pct float64) (float64, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	if pct < svc.Cfg.MinCommissionPct || pct > svc.Cfg.MaxCommissionPct {
		return 0, fmt.Errorf("commission percentage must be between %.1f and %.1f", svc.Cfg.MinCommissionPct, svc.Cfg.MaxCommissionPct)
	}
	svc.Store.CommissionPct = pct
	return svc.Store.CommissionPct, nil
}

// ── RECOMMENDATION SERVICE ──────────────────────────────────────────────────

func (svc *Services) GetRecommendations() map[string]interface{} {
	svc.Store.RLock()
	defer svc.Store.RUnlock()

	var trendingRest []*models.Restaurant
	for _, r := range svc.Store.Restaurants {
		trendingRest = append(trendingRest, r)
	}

	var popularDishes []models.MenuItem
	for _, menuList := range svc.Store.Menus {
		for _, item := range menuList {
			if item.Badge == "Bestseller" || item.Badge == "Chef Special" {
				popularDishes = append(popularDishes, item)
			}
		}
	}

	return map[string]interface{}{
		"trendingRestaurants": trendingRest,
		"popularDishes":       popularDishes,
	}
}

// ── COUPON SERVICE ──────────────────────────────────────────────────────────

func (svc *Services) ValidateCoupon(code string, subtotal float64) (float64, error) {
	svc.Store.RLock()
	defer svc.Store.RUnlock()

	c, exists := svc.Store.Coupons[code]
	if !exists || !c.IsActive {
		return 0, fmt.Errorf("invalid or expired coupon code")
	}

	if subtotal < c.MinOrderEUR {
		return 0, fmt.Errorf("minimum order of €%.2f required for coupon %s", c.MinOrderEUR, code)
	}

	var discount float64
	if c.DiscountEUR != nil {
		discount = *c.DiscountEUR
	} else if c.DiscountPct != nil {
		discount = subtotal * (*c.DiscountPct / 100.0)
	}

	return math.Round(discount*100) / 100, nil
}

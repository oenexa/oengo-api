package models

import "time"

type UserRole string

const (
	RoleCustomer   UserRole = "CUSTOMER"
	RoleRestaurant UserRole = "RESTAURANT"
	RoleCourier    UserRole = "COURIER"
	RoleAdmin      UserRole = "ADMIN"
)

type DigitalWallet struct {
	FiatBalanceEUR   float64 `json:"fiatBalanceEUR"`
	LockedBalanceEUR float64 `json:"lockedBalanceEUR"`
	LoyaltyPoints    int     `json:"loyaltyPoints"`
	CashbackRatePct  float64 `json:"cashbackRatePct"`
}

type User struct {
	ID                  string        `json:"id"`
	Name                string        `json:"name"`
	Email               string        `json:"email"`
	Phone               string        `json:"phone"`
	Role                UserRole      `json:"role"`
	DigitalWallet       DigitalWallet `json:"digitalWallet"`
	CryptoWalletAddress string        `json:"cryptoWalletAddress"`
}

type SavedAddress struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Address   string  `json:"address"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	IsDefault bool    `json:"isDefault"`
	Notes     string  `json:"notes"`
}

type CustomerProfile struct {
	ID                  string         `json:"id"`
	Name                string         `json:"name"`
	Email               string         `json:"email"`
	Phone               string         `json:"phone"`
	Avatar              string         `json:"avatar"`
	SavedAddresses      []SavedAddress `json:"savedAddresses"`
	FavoriteRestaurants []string       `json:"favoriteRestaurants"`
	FavoriteFoods       []string       `json:"favoriteFoods"`
}

type CustomizationOption struct {
	Name     string  `json:"name"`
	ExtraEUR float64 `json:"extraEUR"`
}

type CustomizationGroup struct {
	Name    string                `json:"name"`
	IsMulti bool                  `json:"isMulti"`
	Options []CustomizationOption `json:"options"`
}

type MenuItem struct {
	ID                string               `json:"id"`
	Category          string               `json:"category"`
	Name              string               `json:"name"`
	Description       string               `json:"description"`
	PriceEUR          float64              `json:"priceEUR"`
	PriceOEN          string               `json:"priceOEN"`
	InStock           bool                 `json:"inStock"`
	StockQuantity     int                  `json:"stockQuantity"`
	LowStockThreshold int                  `json:"lowStockThreshold"`
	PrepMinutes       int                  `json:"prepMinutes"`
	Badge             string               `json:"badge"`
	Calories          int                  `json:"calories"`
	Ingredients       string               `json:"ingredients"`
	Customizations    []CustomizationGroup `json:"customizations"`
}

type RestaurantStats struct {
	TotalOrdersCount       int     `json:"totalOrdersCount"`
	ActiveOrdersCount      int     `json:"activeOrdersCount"`
	DeliveredOrdersCount   int     `json:"deliveredOrdersCount"`
	GrossRevenueTodayEUR   float64 `json:"grossRevenueTodayEUR"`
	FiatBalanceEUR         float64 `json:"fiatBalanceEUR"`
	CommissionRetainedPct  float64 `json:"commissionRetainedPct"`
	PlatformCommissionPct  float64 `json:"platformCommissionPct"`
	LegacyLostRevenueEUR   float64 `json:"legacyLostRevenueEUR"`
	AvgPrepTimeMinutes     int     `json:"avgPrepTimeMinutes"`
}

type Restaurant struct {
	ID                  string           `json:"id"`
	Name                string           `json:"name"`
	Tagline             string           `json:"tagline"`
	Cuisine             string           `json:"cuisine"`
	Rating              float64          `json:"rating"`
	ReviewCount         int              `json:"reviewCount"`
	Address             string           `json:"address"`
	Icon                string           `json:"icon"`
	IsOpen              bool             `json:"isOpen"`
	PrepEtaMinutes      int              `json:"prepEtaMinutes"`
	DeliveryFeeEUR      float64          `json:"deliveryFeeEUR"`
	MinOrderEUR         float64          `json:"minOrderEUR"`
	DeliveryRadiusKm    float64          `json:"deliveryRadiusKm"`
	TaxNumber           string           `json:"taxNumber"`
	VATNumber           string           `json:"vatNumber"`
	CryptoWalletAddress string           `json:"cryptoWalletAddress"`
	FiatBalanceEUR      float64          `json:"fiatBalanceEUR"`
	Stats               *RestaurantStats `json:"stats,omitempty"`
}

type CustomizationChoice struct {
	GroupName  string  `json:"groupName"`
	OptionName string  `json:"optionName"`
	ExtraEUR   float64 `json:"extraEUR"`
}

type OrderItem struct {
	ID             string                `json:"id,omitempty"`
	Name           string                `json:"name"`
	Qty            int                   `json:"qty"`
	Price          float64               `json:"price"`
	Customizations []CustomizationChoice `json:"customizations,omitempty"`
	SubtotalEUR    float64               `json:"subtotalEUR,omitempty"`
}

type CardPaymentInfo struct {
	Brand         string `json:"brand"`
	Last4         string `json:"last4"`
	TransactionID string `json:"transactionId"`
}

type OrderStatus string

const (
	StatusAwaitingRestaurant OrderStatus = "AWAITING_RESTAURANT"
	StatusPreparing          OrderStatus = "PREPARING"
	StatusReadyForPickup     OrderStatus = "READY_FOR_PICKUP"
	StatusInTransit          OrderStatus = "IN_TRANSIT"
	StatusDelivered          OrderStatus = "DELIVERED"
	StatusCancelled          OrderStatus = "CANCELLED"
	StatusRefunded           OrderStatus = "REFUNDED"
)

type Order struct {
	ID                 string           `json:"id"`
	BuyerID            string           `json:"buyerId"`
	BuyerName          string           `json:"buyerName"`
	BuyerAddress       string           `json:"buyerAddress"`
	RestaurantID       string           `json:"restaurantId"`
	RestaurantAddress  string           `json:"restaurantAddress"`
	CourierID          string           `json:"courierId,omitempty"`
	CourierAddress     string           `json:"courierAddress,omitempty"`
	Items              []OrderItem      `json:"items"`
	Amount             float64          `json:"amount"`
	DeliveryFee        float64          `json:"deliveryFee"`
	Tip                float64          `json:"tip"`
	DiscountEUR        float64          `json:"discountEUR"`
	Total              float64          `json:"total"`
	CommissionPct      float64          `json:"commissionPct"`
	PaymentMethod      string           `json:"paymentMethod"`
	CardPayment        *CardPaymentInfo `json:"cardPayment,omitempty"`
	PaymentStatus      string           `json:"paymentStatus"`
	Status             OrderStatus      `json:"status"`
	PickupBarcode      string           `json:"pickupBarcode"`
	PickupBarcodeHash  string           `json:"pickupBarcodeHash"`
	DeliveryPIN        string           `json:"deliveryPin"`
	DeliveryPINHash    string           `json:"deliveryPinHash"`
	CreatedAt          string           `json:"createdAt"`
	DeliveredAt        string           `json:"deliveredAt,omitempty"`
	EscrowLocked       bool             `json:"escrowLocked"`
	RestaurantPayout   float64          `json:"restaurantPayout,omitempty"`
	CourierPayout      float64          `json:"courierPayout,omitempty"`
}

type LedgerEntryType string

const (
	LedgerDebit  LedgerEntryType = "DEBIT"
	LedgerCredit LedgerEntryType = "CREDIT"
)

type LedgerEntry struct {
	ID          string          `json:"id"`
	TxID        string          `json:"txId"`
	AccountCode string          `json:"accountCode"` // 1001, 1002, 1003, 2001, 2002, 3001, 3002
	EntryType   LedgerEntryType `json:"entryType"`
	AmountEUR   float64         `json:"amountEUR"`
	Description string          `json:"description"`
	CreatedAt   string          `json:"createdAt"`
}

type CoinTx struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // EARNED, REDEEMED
	Coins     int64     `json:"coins"`
	Reason    string    `json:"reason"`
	OrderID   string    `json:"orderId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type CoinProfile struct {
	UserID       string   `json:"userId"`
	CoinBalance  int64    `json:"coinBalance"`
	TotalEarned  int64    `json:"totalEarned"`
	TotalSpent   int64    `json:"totalSpent"`
	Transactions []CoinTx `json:"transactions"`
}

type Location struct {
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Heading   float64 `json:"heading"`
	UpdatedAt string  `json:"updatedAt"`
}

type RiderProfile struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Phone            string   `json:"phone"`
	IsOnline         bool     `json:"isOnline"`
	VehicleType      string   `json:"vehicleType"`
	VehiclePlate     string   `json:"vehiclePlate"`
	KYCStatus        string   `json:"kycStatus"`
	Rating           float64  `json:"rating"`
	TotalDeliveries  int      `json:"totalDeliveries"`
	TodayEarningsEUR float64  `json:"todayEarningsEUR"`
	CurrentLocation  Location `json:"currentLocation"`
}

type Coupon struct {
	Code         string   `json:"code"`
	DiscountEUR  *float64 `json:"discountEUR,omitempty"`
	DiscountPct  *float64 `json:"discountPct,omitempty"`
	MinOrderEUR  float64  `json:"minOrderEUR"`
	MaxUses      int      `json:"maxUses"`
	TimesUsed    int      `json:"timesUsed"`
	IsActive     bool     `json:"isActive"`
}

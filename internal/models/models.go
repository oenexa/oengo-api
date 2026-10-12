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
	FiatBalanceEUR   float64 `json:"fiatBalanceEUR" gorm:"default:0"`
	LockedBalanceEUR float64 `json:"lockedBalanceEUR" gorm:"default:0"`
	LoyaltyPoints    int     `json:"loyaltyPoints" gorm:"default:0"`
	CashbackRatePct  float64 `json:"cashbackRatePct" gorm:"default:0"`
}

type User struct {
	ID                  string        `json:"id" gorm:"primaryKey;type:varchar(100)"`
	Name                string        `json:"name" gorm:"type:varchar(255)"`
	Email               string        `json:"email" gorm:"type:varchar(255);uniqueIndex"`
	Phone               string        `json:"phone" gorm:"type:varchar(50)"`
	PasswordHash        string        `json:"-" gorm:"type:varchar(255)"`
	Role                UserRole      `json:"role" gorm:"type:varchar(50)"`
	DigitalWallet       DigitalWallet `json:"digitalWallet" gorm:"embedded"`
	CryptoWalletAddress string        `json:"cryptoWalletAddress" gorm:"type:varchar(100)"`
	KYCStatus           string        `json:"kycStatus,omitempty" gorm:"type:varchar(50)"`
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
	ID                  string         `json:"id" gorm:"primaryKey;type:varchar(100)"`
	Name                string         `json:"name" gorm:"type:varchar(255)"`
	Email               string         `json:"email" gorm:"type:varchar(255)"`
	Phone               string         `json:"phone" gorm:"type:varchar(50)"`
	Avatar              string         `json:"avatar" gorm:"type:varchar(255)"`
	SavedAddresses      []SavedAddress `json:"savedAddresses" gorm:"serializer:json"`
	FavoriteRestaurants []string       `json:"favoriteRestaurants" gorm:"serializer:json"`
	FavoriteFoods       []string       `json:"favoriteFoods" gorm:"serializer:json"`
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
	ID                string               `json:"id" gorm:"primaryKey;type:varchar(100)"`
	RestaurantID      string               `json:"restaurantId,omitempty" gorm:"type:varchar(100);index"`
	Category          string               `json:"category" gorm:"type:varchar(100)"`
	Name              string               `json:"name" gorm:"type:varchar(255);not null"`
	Description       string               `json:"description" gorm:"type:text"`
	PriceEUR          float64              `json:"priceEUR" gorm:"type:numeric(10,2);not null"`
	PriceOEN          string               `json:"priceOEN" gorm:"type:varchar(50)"`
	InStock           bool                 `json:"inStock" gorm:"default:true"`
	StockQuantity     int                  `json:"stockQuantity" gorm:"default:20"`
	LowStockThreshold int                  `json:"lowStockThreshold" gorm:"default:5"`
	PrepMinutes       int                  `json:"prepMinutes" gorm:"default:15"`
	Badge             string               `json:"badge" gorm:"type:varchar(50)"`
	Calories          int                  `json:"calories" gorm:"default:0"`
	Ingredients       string               `json:"ingredients" gorm:"type:text"`
	Customizations    []CustomizationGroup `json:"customizations" gorm:"serializer:json"`
}

type RestaurantStats struct {
	TotalOrdersCount      int     `json:"totalOrdersCount"`
	ActiveOrdersCount     int     `json:"activeOrdersCount"`
	DeliveredOrdersCount  int     `json:"deliveredOrdersCount"`
	GrossRevenueTodayEUR  float64 `json:"grossRevenueTodayEUR"`
	FiatBalanceEUR        float64 `json:"fiatBalanceEUR"`
	CommissionRetainedPct float64 `json:"commissionRetainedPct"`
	PlatformCommissionPct float64 `json:"platformCommissionPct"`
	LegacyLostRevenueEUR  float64 `json:"legacyLostRevenueEUR"`
	AvgPrepTimeMinutes    int     `json:"avgPrepTimeMinutes"`
}

type Restaurant struct {
	ID                  string           `json:"id" gorm:"primaryKey;type:varchar(100)"`
	Name                string           `json:"name" gorm:"type:varchar(255);not null"`
	Tagline             string           `json:"tagline" gorm:"type:varchar(255)"`
	Cuisine             string           `json:"cuisine" gorm:"type:varchar(100);index"`
	Rating              float64          `json:"rating" gorm:"default:4.8"`
	ReviewCount         int              `json:"reviewCount" gorm:"default:0"`
	Address             string           `json:"address" gorm:"type:varchar(255)"`
	Icon                string           `json:"icon" gorm:"type:varchar(50)"`
	IsOpen              bool             `json:"isOpen" gorm:"default:true"`
	PrepEtaMinutes      int              `json:"prepEtaMinutes" gorm:"default:20"`
	DeliveryFeeEUR      float64          `json:"deliveryFeeEUR" gorm:"default:2.50"`
	MinOrderEUR         float64          `json:"minOrderEUR" gorm:"default:10.00"`
	DeliveryRadiusKm    float64          `json:"deliveryRadiusKm" gorm:"default:5.0"`
	TaxNumber           string           `json:"taxNumber" gorm:"type:varchar(50)"`
	VATNumber           string           `json:"vatNumber" gorm:"type:varchar(50)"`
	CryptoWalletAddress string           `json:"cryptoWalletAddress" gorm:"type:varchar(100)"`
	FiatBalanceEUR      float64          `json:"fiatBalanceEUR" gorm:"default:0"`
	Stats               *RestaurantStats `json:"stats,omitempty" gorm:"-"`
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
	ID                string           `json:"id" gorm:"primaryKey;type:varchar(100)"`
	BuyerID           string           `json:"buyerId" gorm:"type:varchar(100);index"`
	BuyerName         string           `json:"buyerName" gorm:"type:varchar(255)"`
	BuyerAddress      string           `json:"buyerAddress" gorm:"type:varchar(255)"`
	RestaurantID      string           `json:"restaurantId" gorm:"type:varchar(100);index"`
	RestaurantAddress string           `json:"restaurantAddress" gorm:"type:varchar(255)"`
	CourierID         string           `json:"courierId,omitempty" gorm:"type:varchar(100);index"`
	CourierAddress    string           `json:"courierAddress,omitempty" gorm:"type:varchar(255)"`
	Items             []OrderItem      `json:"items" gorm:"serializer:json"`
	Amount            float64          `json:"amount" gorm:"type:numeric(10,2);not null"`
	DeliveryFee       float64          `json:"deliveryFee" gorm:"type:numeric(10,2)"`
	Tip               float64          `json:"tip" gorm:"type:numeric(10,2)"`
	DiscountEUR       float64          `json:"discountEUR" gorm:"type:numeric(10,2)"`
	Total             float64          `json:"total" gorm:"type:numeric(10,2);not null"`
	CommissionPct     float64          `json:"commissionPct" gorm:"type:numeric(5,2)"`
	PaymentMethod     string           `json:"paymentMethod" gorm:"type:varchar(50)"`
	CardPayment       *CardPaymentInfo `json:"cardPayment,omitempty" gorm:"serializer:json"`
	PaymentStatus     string           `json:"paymentStatus" gorm:"type:varchar(50)"`
	Status            OrderStatus      `json:"status" gorm:"type:varchar(50);index"`
	PickupBarcode     string           `json:"pickupBarcode" gorm:"type:varchar(100)"`
	PickupBarcodeHash string           `json:"pickupBarcodeHash" gorm:"type:varchar(100)"`
	DeliveryPIN       string           `json:"deliveryPin" gorm:"type:varchar(10)"`
	DeliveryPINHash   string           `json:"deliveryPinHash" gorm:"type:varchar(100)"`
	CreatedAt         string           `json:"createdAt" gorm:"type:varchar(100)"`
	DeliveredAt       string           `json:"deliveredAt,omitempty" gorm:"type:varchar(100)"`
	EscrowLocked      bool             `json:"escrowLocked" gorm:"default:true"`
	RestaurantPayout  float64          `json:"restaurantPayout,omitempty" gorm:"type:numeric(10,2)"`
	CourierPayout     float64          `json:"courierPayout,omitempty" gorm:"type:numeric(10,2)"`
}

type LedgerEntryType string

const (
	LedgerDebit  LedgerEntryType = "DEBIT"
	LedgerCredit LedgerEntryType = "CREDIT"
)

type LedgerEntry struct {
	ID          string          `json:"id" gorm:"primaryKey;type:varchar(100)"`
	TxID        string          `json:"txId" gorm:"type:varchar(100);index"`
	AccountCode string          `json:"accountCode" gorm:"type:varchar(50);index"`
	EntryType   LedgerEntryType `json:"entryType" gorm:"type:varchar(20)"`
	AmountEUR   float64         `json:"amountEUR" gorm:"type:numeric(10,2);not null"`
	Description string          `json:"description" gorm:"type:varchar(255)"`
	CreatedAt   string          `json:"createdAt" gorm:"type:varchar(100)"`
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
	UserID       string   `json:"userId" gorm:"primaryKey;type:varchar(100)"`
	CoinBalance  int64    `json:"coinBalance" gorm:"default:0"`
	TotalEarned  int64    `json:"totalEarned" gorm:"default:0"`
	TotalSpent   int64    `json:"totalSpent" gorm:"default:0"`
	Transactions []CoinTx `json:"transactions" gorm:"serializer:json"`
}

type Location struct {
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Heading   float64 `json:"heading"`
	UpdatedAt string  `json:"updatedAt"`
}

type RiderProfile struct {
	ID               string   `json:"id" gorm:"primaryKey;type:varchar(100)"`
	Name             string   `json:"name" gorm:"type:varchar(255)"`
	Phone            string   `json:"phone" gorm:"type:varchar(50)"`
	IsOnline         bool     `json:"isOnline" gorm:"default:false"`
	VehicleType      string   `json:"vehicleType" gorm:"type:varchar(50)"`
	VehiclePlate     string   `json:"vehiclePlate" gorm:"type:varchar(50)"`
	KYCStatus        string   `json:"kycStatus" gorm:"type:varchar(50)"`
	Rating           float64  `json:"rating" gorm:"default:4.9"`
	TotalDeliveries  int      `json:"totalDeliveries" gorm:"default:0"`
	TodayEarningsEUR float64  `json:"todayEarningsEUR" gorm:"default:0"`
	CurrentLocation  Location `json:"currentLocation" gorm:"serializer:json"`
}

type Coupon struct {
	Code        string   `json:"code" gorm:"primaryKey;type:varchar(50)"`
	DiscountEUR *float64 `json:"discountEUR,omitempty" gorm:"type:numeric(10,2)"`
	DiscountPct *float64 `json:"discountPct,omitempty" gorm:"type:numeric(5,2)"`
	MinOrderEUR float64  `json:"minOrderEUR" gorm:"type:numeric(10,2);default:0"`
	MaxUses     int      `json:"maxUses" gorm:"default:0"`
	TimesUsed   int      `json:"timesUsed" gorm:"default:0"`
	IsActive    bool     `json:"isActive" gorm:"default:true"`
}

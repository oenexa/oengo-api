package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Banking Standard Base Model replacing gorm.Model
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `gorm:"not null;default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"not null;default:current_timestamp" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// -----------------------------------------------------------------------------
// CORE DOMAIN
// -----------------------------------------------------------------------------

// BankingUser represents the strict user identity
type BankingUser struct {
	BaseModel
	Name         string   `gorm:"type:varchar(255);not null" json:"name"`
	Email        string   `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Phone        string   `gorm:"type:varchar(50)" json:"phone"`
	PasswordHash string   `gorm:"type:varchar(255);not null" json:"-"`
	Role         UserRole `gorm:"type:varchar(20);not null" json:"role"`
	KYCStatus    string   `gorm:"type:varchar(20);default:'PENDING'" json:"kycStatus"`

	// Relationships
	Wallets      []BankingWallet `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"wallets"`
}

// -----------------------------------------------------------------------------
// FINANCIAL DOMAIN (BANKING STANDARD)
// -----------------------------------------------------------------------------

// WalletStatus limits allowed operations
type WalletStatus string
const (
	WalletActive  WalletStatus = "ACTIVE"
	WalletFrozen  WalletStatus = "FROZEN"
	WalletClosed  WalletStatus = "CLOSED"
)

// BankingWallet acts as a financial container
type BankingWallet struct {
	BaseModel
	UserID    uuid.UUID    `gorm:"type:uuid;not null;index" json:"userId"`
	Type      string       `gorm:"type:varchar(50);not null" json:"type"` // e.g., 'FIAT_EUR', 'OEN_TOKEN'
	Status    WalletStatus `gorm:"type:varchar(20);not null;default:'ACTIVE'" json:"status"`
	
	// Balance is the CACHED sum of LedgerEntries. 
	// True balance is always calculated from the immutable Ledger.
	Balance   decimal.Decimal `gorm:"type:numeric(28,8);not null;default:0.00000000;check:balance >= 0" json:"balance"`
	Currency  string          `gorm:"type:varchar(5);not null;default:'EUR'" json:"currency"`

	// Double-entry account mapping
	AccountID uuid.UUID       `gorm:"type:uuid;not null" json:"accountId"`
}

// -----------------------------------------------------------------------------
// IMMUTABLE LEDGER (DOUBLE-ENTRY BOOKKEEPING)
// -----------------------------------------------------------------------------

type LedgerAccountType string
const (
	AccountAsset     LedgerAccountType = "ASSET"     // User Wallets
	AccountLiability LedgerAccountType = "LIABILITY" // Escrow, Outstanding
	AccountEquity    LedgerAccountType = "EQUITY"    // Retained Earnings
	AccountRevenue   LedgerAccountType = "REVENUE"   // Commissions collected
	AccountExpense   LedgerAccountType = "EXPENSE"   // Platform costs
)

// LedgerAccount is the chart of accounts
type LedgerAccount struct {
	BaseModel
	Name     string            `gorm:"type:varchar(100);not null;unique" json:"name"`
	Type     LedgerAccountType `gorm:"type:varchar(20);not null" json:"type"`
	Currency string            `gorm:"type:varchar(5);not null;default:'EUR'" json:"currency"`
}

// LedgerTransaction groups multiple entries (debits/credits) that must balance to zero
type LedgerTransaction struct {
	BaseModel
	IdempotencyKey string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"idempotencyKey"` // Prevents double-billing
	ReferenceID   uuid.UUID `gorm:"type:uuid;index" json:"referenceId"` // Maps to OrderID or external reference
	ReferenceType string    `gorm:"type:varchar(50)" json:"referenceType"`
	Description   string    `gorm:"type:varchar(255)" json:"description"`
	Status        string    `gorm:"type:varchar(20);not null;default:'POSTED'" json:"status"`
	
	// Cryptographic Audit Trail (Blockchain concept)
	PreviousHash  string    `gorm:"type:varchar(64);not null" json:"previousHash"`
	Hash          string    `gorm:"type:varchar(64);not null" json:"hash"`

	Entries       []LedgerEntryRecord `gorm:"foreignKey:TransactionID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"entries"`
}

func (t *LedgerTransaction) BeforeUpdate(tx *gorm.DB) (err error) {
	return errors.New("LedgerTransaction is immutable and cannot be updated")
}
func (t *LedgerTransaction) BeforeDelete(tx *gorm.DB) (err error) {
	return errors.New("LedgerTransaction is immutable and cannot be deleted")
}

type EntryDirection string
const (
	DirectionDebit  EntryDirection = "DEBIT"
	DirectionCredit EntryDirection = "CREDIT"
)

// LedgerEntryRecord is the immutable audit line-item
type LedgerEntryRecord struct {
	BaseModel
	TransactionID uuid.UUID       `gorm:"type:uuid;not null;index" json:"transactionId"`
	AccountID     uuid.UUID       `gorm:"type:uuid;not null;index" json:"accountId"`
	Amount        decimal.Decimal `gorm:"type:numeric(28,8);not null" json:"amount"`
	Direction     EntryDirection  `gorm:"type:varchar(10);not null" json:"direction"`
	Currency      string          `gorm:"type:varchar(5);not null;default:'EUR'" json:"currency"`
}

func (e *LedgerEntryRecord) BeforeUpdate(tx *gorm.DB) (err error) {
	return errors.New("LedgerEntryRecord is immutable and cannot be updated")
}
func (e *LedgerEntryRecord) BeforeDelete(tx *gorm.DB) (err error) {
	return errors.New("LedgerEntryRecord is immutable and cannot be deleted")
}

// -----------------------------------------------------------------------------
// DOMAIN EXTENSIONS
// -----------------------------------------------------------------------------

// BankingOrder represents a locked financial agreement
type BankingOrder struct {
	BaseModel
	BuyerID       uuid.UUID       `gorm:"type:uuid;not null;index" json:"buyerId"`
	RestaurantID  uuid.UUID       `gorm:"type:uuid;not null;index" json:"restaurantId"`
	CourierID     *uuid.UUID      `gorm:"type:uuid;index" json:"courierId,omitempty"` // nullable
	
	Status        OrderStatus     `gorm:"type:varchar(30);not null;default:'PENDING'" json:"status"`
	
	Subtotal      decimal.Decimal `gorm:"type:numeric(28,8);not null" json:"subtotal"`
	DeliveryFee   decimal.Decimal `gorm:"type:numeric(28,8);not null" json:"deliveryFee"`
	Tip           decimal.Decimal `gorm:"type:numeric(28,8);not null;default:0" json:"tip"`
	PlatformFee   decimal.Decimal `gorm:"type:numeric(28,8);not null" json:"platformFee"`
	TotalAmount   decimal.Decimal `gorm:"type:numeric(28,8);not null" json:"totalAmount"`
	
	Currency      string          `gorm:"type:varchar(5);not null;default:'EUR'" json:"currency"`
	PaymentMethod string          `gorm:"type:varchar(50);not null" json:"paymentMethod"`
}

package repository

import "oengo-api/internal/models"

// UserRepository defines the database operations for Users.
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(user *models.User) error
}

// OrderRepository defines the database operations for Orders.
type OrderRepository interface {
	CreateOrder(order *models.Order) error
	GetOrderByID(id string) (*models.Order, error)
	UpdateOrderStatus(id string, status models.OrderStatus) error
}

// WalletRepository defines operations for the Double-Entry Ledger and Wallets.
type WalletRepository interface {
	GetWalletByUserID(userID string) (*models.DigitalWallet, error)
	Transact(entries []models.LedgerEntry) error
}

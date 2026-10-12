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
	ListOrders(restaurantID, buyerID, status string) ([]*models.Order, error)
	UpdateOrderStatus(id string, status models.OrderStatus) error
	UpdateOrder(order *models.Order) error
}

// RestaurantRepository defines database operations for Merchants and Menus.
type RestaurantRepository interface {
	ListRestaurants(cuisine, search string) ([]*models.Restaurant, error)
	GetRestaurantByID(id string) (*models.Restaurant, error)
	CreateRestaurant(r *models.Restaurant) error
	UpdateRestaurant(r *models.Restaurant) error
	GetMenu(restaurantID string) ([]models.MenuItem, error)
	AddDish(restaurantID string, item *models.MenuItem) error
	UpdateDish(restaurantID string, item *models.MenuItem) error
	DeleteDish(restaurantID, itemID string) error
}

// WalletRepository defines operations for the Double-Entry Ledger and Wallets.
type WalletRepository interface {
	GetWalletByUserID(userID string) (*models.DigitalWallet, error)
	UpdateWallet(userID string, wallet *models.DigitalWallet) error
	Transact(entries []models.LedgerEntry) error
	GetLedgerEntries(limit int) ([]models.LedgerEntry, error)
}

// RiderRepository defines operations for Couriers and Dispatch.
type RiderRepository interface {
	GetRiderByID(id string) (*models.RiderProfile, error)
	SaveRider(rider *models.RiderProfile) error
	ListOnlineRiders() ([]*models.RiderProfile, error)
}

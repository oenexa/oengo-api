package repository

import (
	"errors"

	"oengo-api/internal/apperrors"
	"oengo-api/internal/models"
	"gorm.io/gorm"
)

type PostgresOrderRepository struct {
	db *gorm.DB
}

func NewPostgresOrderRepository(db *gorm.DB) *PostgresOrderRepository {
	db.AutoMigrate(&models.Order{})
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) CreateOrder(order *models.Order) error {
	return apperrors.WrapDBError("CreateOrder", r.db.Create(order).Error)
}

func (r *PostgresOrderRepository) GetOrderByID(id string) (*models.Order, error) {
	var order models.Order
	result := r.db.Where("id = ?", id).First(&order)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &order, apperrors.WrapDBError("GetOrderByID", result.Error)
}

func (r *PostgresOrderRepository) ListOrders(restaurantID, buyerID, status string) ([]*models.Order, error) {
	query := r.db.Model(&models.Order{})
	if restaurantID != "" {
		query = query.Where("restaurant_id = ?", restaurantID)
	}
	if buyerID != "" {
		query = query.Where("buyer_id = ?", buyerID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var orders []*models.Order
	err := query.Order("created_at desc").Find(&orders).Error
	return orders, apperrors.WrapDBError("ListOrders", err)
}

func (r *PostgresOrderRepository) UpdateOrderStatus(id string, status models.OrderStatus) error {
	return apperrors.WrapDBError("UpdateOrderStatus", r.db.Model(&models.Order{}).Where("id = ?", id).Update("status", status).Error)
}

func (r *PostgresOrderRepository) UpdateOrder(order *models.Order) error {
	return apperrors.WrapDBError("UpdateOrder", r.db.Save(order).Error)
}

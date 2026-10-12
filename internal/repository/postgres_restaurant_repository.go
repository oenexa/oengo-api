package repository

import (
	"errors"
	"strings"

	"oengo-api/internal/apperrors"
	"oengo-api/internal/models"
	"gorm.io/gorm"
)

type PostgresRestaurantRepository struct {
	db *gorm.DB
}

func NewPostgresRestaurantRepository(db *gorm.DB) *PostgresRestaurantRepository {
	db.AutoMigrate(&models.Restaurant{}, &models.MenuItem{})
	return &PostgresRestaurantRepository{db: db}
}

func (r *PostgresRestaurantRepository) ListRestaurants(cuisine, search string) ([]*models.Restaurant, error) {
	query := r.db.Model(&models.Restaurant{})

	if cuisine != "" && cuisine != "ALL" {
		query = query.Where("LOWER(cuisine) LIKE ?", "%"+strings.ToLower(cuisine)+"%")
	}
	if search != "" {
		s := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(cuisine) LIKE ? OR LOWER(tagline) LIKE ?", s, s, s)
	}

	var list []*models.Restaurant
	err := query.Find(&list).Error
	return list, apperrors.WrapDBError("ListRestaurants", err)
}

func (r *PostgresRestaurantRepository) GetRestaurantByID(id string) (*models.Restaurant, error) {
	var rest models.Restaurant
	result := r.db.Where("id = ?", id).First(&rest)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &rest, apperrors.WrapDBError("GetRestaurantByID", result.Error)
}

func (r *PostgresRestaurantRepository) CreateRestaurant(rest *models.Restaurant) error {
	return apperrors.WrapDBError("CreateRestaurant", r.db.Create(rest).Error)
}

func (r *PostgresRestaurantRepository) UpdateRestaurant(rest *models.Restaurant) error {
	return apperrors.WrapDBError("UpdateRestaurant", r.db.Save(rest).Error)
}

func (r *PostgresRestaurantRepository) GetMenu(restaurantID string) ([]models.MenuItem, error) {
	var items []models.MenuItem
	err := r.db.Where("restaurant_id = ?", restaurantID).Find(&items).Error
	return items, apperrors.WrapDBError("GetMenu", err)
}

func (r *PostgresRestaurantRepository) AddDish(restaurantID string, item *models.MenuItem) error {
	item.RestaurantID = restaurantID
	return apperrors.WrapDBError("AddDish", r.db.Create(item).Error)
}

func (r *PostgresRestaurantRepository) UpdateDish(restaurantID string, item *models.MenuItem) error {
	item.RestaurantID = restaurantID
	return apperrors.WrapDBError("UpdateDish", r.db.Save(item).Error)
}

func (r *PostgresRestaurantRepository) DeleteDish(restaurantID, itemID string) error {
	err := r.db.Where("id = ? AND restaurant_id = ?", itemID, restaurantID).Delete(&models.MenuItem{}).Error
	return apperrors.WrapDBError("DeleteDish", err)
}

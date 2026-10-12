package repository

import (
	"errors"

	"oengo-api/internal/models"
	"gorm.io/gorm"
)

type PostgresRiderRepository struct {
	db *gorm.DB
}

func NewPostgresRiderRepository(db *gorm.DB) *PostgresRiderRepository {
	db.AutoMigrate(&models.RiderProfile{})
	return &PostgresRiderRepository{db: db}
}

func (r *PostgresRiderRepository) GetRiderByID(id string) (*models.RiderProfile, error) {
	var rider models.RiderProfile
	result := r.db.Where("id = ?", id).First(&rider)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &rider, result.Error
}

func (r *PostgresRiderRepository) SaveRider(rider *models.RiderProfile) error {
	return r.db.Save(rider).Error
}

func (r *PostgresRiderRepository) ListOnlineRiders() ([]*models.RiderProfile, error) {
	var riders []*models.RiderProfile
	err := r.db.Where("is_online = ?", true).Find(&riders).Error
	return riders, err
}

package repository

import (
	"errors"

	"oengo-api/internal/apperrors"
	"oengo-api/internal/models"
	"gorm.io/gorm"
)

type PostgresUserRepository struct {
	db *gorm.DB
}

func NewPostgresUserRepository(db *gorm.DB) *PostgresUserRepository {
	// AutoMigrate will ensure the database schema matches the struct
	db.AutoMigrate(&models.User{})
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(user *models.User) error {
	result := r.db.Create(user)
	return apperrors.WrapDBError("CreateUser", result.Error)
}

func (r *PostgresUserRepository) GetUserByID(id string) (*models.User, error) {
	var user models.User
	result := r.db.Where("id = ?", id).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil // Not found
	}
	return &user, apperrors.WrapDBError("GetUserByID", result.Error)
}

func (r *PostgresUserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, apperrors.WrapDBError("GetUserByEmail", result.Error)
}

func (r *PostgresUserRepository) UpdateUser(user *models.User) error {
	result := r.db.Save(user)
	return apperrors.WrapDBError("UpdateUser", result.Error)
}

package services

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"oengo-api/internal/models"
)

type RegisterRequest struct {
	Name     string          `json:"name"`
	Email    string          `json:"email"`
	Phone    string          `json:"phone"`
	Password string          `json:"password"`
	Role     models.UserRole `json:"role"` // CUSTOMER, RESTAURANT, COURIER, ADMIN
}

func (svc *Services) RegisterUser(req RegisterRequest) (*models.BankingUser, error) {
	// 1. Check if email exists
	var existing models.BankingUser
	if err := svc.Store.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("user with email already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 2. Hash Password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	kycStatus := "PENDING"
	if req.Role == models.RoleCustomer || req.Role == models.RoleAdmin {
		kycStatus = "VERIFIED"
	}

	user := &models.BankingUser{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hash),
		Role:         req.Role,
		KYCStatus:    kycStatus,
	}

	// 3. Save to Postgres
	if err := svc.Store.DB.Create(user).Error; err != nil {
		return nil, err
	}

	// (Legacy support fallback to keep memory maps working during transition)
	svc.Store.Lock()
	svc.Store.Users[user.ID.String()] = &models.User{
		ID:           user.ID.String(),
		Name:         user.Name,
		Email:        user.Email,
		Role:         user.Role,
	}
	svc.Store.Unlock()

	return user, nil
}

func (svc *Services) LoginUser(email, password string) (*models.BankingUser, error) {
	var user models.BankingUser
	if err := svc.Store.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	
	return &user, nil
}

func (svc *Services) ApproveKYC(userID string) (*models.User, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	u, exists := svc.Store.Users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	u.KYCStatus = "VERIFIED"

	if u.Role == models.RoleCourier {
		if rp, ok := svc.Store.Riders[userID]; ok {
			rp.KYCStatus = "VERIFIED"
		}
	}
	return u, nil
}

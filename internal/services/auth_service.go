package services

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"oengo-api/internal/models"
)

type RegisterRequest struct {
	Name     string          `json:"name"`
	Email    string          `json:"email"`
	Phone    string          `json:"phone"`
	Password string          `json:"password"`
	Role     models.UserRole `json:"role"` // CUSTOMER, RESTAURANT, COURIER, ADMIN
}

func (svc *Services) RegisterUser(req RegisterRequest) (*models.User, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	// Check if email already exists
	for _, u := range svc.Store.Users {
		if u.Email == req.Email {
			return nil, fmt.Errorf("user with email already exists")
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID := fmt.Sprintf("usr_%d", time.Now().UnixMilli())
	
	// Initial KYC status
	kycStatus := "PENDING"
	if req.Role == models.RoleCustomer || req.Role == models.RoleAdmin {
		kycStatus = "VERIFIED" // Customers/Admins auto-verified for now
	}

	u := &models.User{
		ID:           userID,
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hash),
		Role:         req.Role,
		KYCStatus:    kycStatus,
		DigitalWallet: models.DigitalWallet{
			FiatBalanceEUR: 0.0,
		},
	}
	svc.Store.Users[userID] = u

	// Initialize corresponding profiles based on role
	if req.Role == models.RoleCustomer {
		svc.Store.CustomerProfiles[userID] = &models.CustomerProfile{
			ID:     userID,
			Name:   req.Name,
			Email:  req.Email,
			Phone:  req.Phone,
			Avatar: "👤",
		}
		svc.Store.Coins[userID] = &models.CoinProfile{
			UserID:      userID,
			CoinBalance: 0,
		}
	} else if req.Role == models.RoleRestaurant {
		svc.Store.Restaurants[userID] = &models.Restaurant{
			ID:      userID,
			Name:    req.Name,
			Address: "Update Address in Portal",
			IsOpen:  false,
		}
	} else if req.Role == models.RoleCourier {
		svc.Store.Riders[userID] = &models.RiderProfile{
			ID:        userID,
			Name:      req.Name,
			Phone:     req.Phone,
			KYCStatus: kycStatus,
		}
	}

	return u, nil
}

func (svc *Services) LoginUser(email, password string) (*models.User, error) {
	svc.Store.RLock()
	defer svc.Store.RUnlock()

	for _, u := range svc.Store.Users {
		if u.Email == email {
			err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
			if err != nil {
				return nil, fmt.Errorf("invalid credentials")
			}
			return u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
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

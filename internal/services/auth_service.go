package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"oengo-api/internal/models"
)

type RegisterRequest struct {
	Name            string          `json:"name"`
	Email           string          `json:"email"`
	Phone           string          `json:"phone"`
	Password        string          `json:"password"`
	Role            models.UserRole `json:"role"` // CUSTOMER, RESTAURANT, COURIER, ADMIN
	BusinessName    string          `json:"businessName,omitempty"`
	VatID           string          `json:"vatId,omitempty"`
	IBAN            string          `json:"iban,omitempty"`
	BIC             string          `json:"bic,omitempty"`
	BusinessAddress string          `json:"businessAddress,omitempty"`
	Country         string          `json:"country,omitempty"`
	GDPRConsent     bool            `json:"gdprConsent,omitempty"`
}

func (svc *Services) RegisterUser(req RegisterRequest) (*models.BankingUser, error) {
	// 1. Check if email exists
	if svc.Store.DB != nil {
		var existing models.BankingUser
		if err := svc.Store.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
			return nil, fmt.Errorf("user with email already exists")
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	} else {
		svc.Store.RLock()
		for _, u := range svc.Store.Users {
			if u.Email == req.Email {
				svc.Store.RUnlock()
				return nil, fmt.Errorf("user with email already exists")
			}
		}
		svc.Store.RUnlock()
	}

	// 2. Hash Password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Under EU Anti-Money Laundering (AMLD5) & PSD2 regulations:
	// Business accounts (RESTAURANT) and financial participants (COURIER) MUST be verified via KYC.
	kycStatus := "PENDING"
	if req.Role == models.RoleCustomer || req.Role == models.RoleAdmin {
		kycStatus = "VERIFIED"
	}

	var consentTime *time.Time
	if req.GDPRConsent {
		now := time.Now().UTC()
		consentTime = &now
	}

	country := req.Country
	if country == "" {
		country = "IT"
	}

	user := &models.BankingUser{
		Name:            req.Name,
		Email:           req.Email,
		Phone:           req.Phone,
		PasswordHash:    string(hash),
		Role:            req.Role,
		KYCStatus:       kycStatus,
		BusinessName:    req.BusinessName,
		VatID:           req.VatID,
		IBAN:            req.IBAN,
		BIC:             req.BIC,
		BusinessAddress: req.BusinessAddress,
		Country:         country,
		GDPRConsentAt:   consentTime,
	}

	// 3. Save to Postgres if DB is connected
	if svc.Store.DB != nil {
		if err := svc.Store.DB.Create(user).Error; err != nil {
			return nil, err
		}
	} else {
		// Mock UUID for testing
		user.ID = uuid.New()
	}

	// (Synchronize with memory map)
	svc.Store.Lock()
	svc.Store.Users[user.ID.String()] = &models.User{
		ID:           user.ID.String(),
		Name:         user.Name,
		Email:        user.Email,
		Phone:        user.Phone,
		PasswordHash: user.PasswordHash,
		Role:         user.Role,
		KYCStatus:    user.KYCStatus,
	}
	if user.Role == models.RoleCourier {
		svc.Store.Riders[user.ID.String()] = &models.RiderProfile{
			ID:        user.ID.String(),
			Name:      user.Name,
			Phone:     user.Phone,
			KYCStatus: user.KYCStatus,
			IsOnline:  true,
		}
	}
	svc.Store.Unlock()

	return user, nil
}

func (svc *Services) LoginUser(email, password string) (*models.BankingUser, error) {
	if svc.Store.DB != nil {
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

	// In-memory fallback
	svc.Store.RLock()
	defer svc.Store.RUnlock()
	for _, u := range svc.Store.Users {
		if u.Email == email {
			err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
			if err != nil {
				return nil, fmt.Errorf("invalid credentials")
			}
			uid, _ := uuid.Parse(u.ID)
			return &models.BankingUser{
				BaseModel: models.BaseModel{ID: uid},
				Name:      u.Name,
				Email:     u.Email,
				Phone:     u.Phone,
				Role:      u.Role,
				KYCStatus: u.KYCStatus,
			}, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// LoginOrRegisterOAuth supports seamless single-click Google/Gmail and Passkey authentication
func (svc *Services) LoginOrRegisterOAuth(name, email, provider, providerID string, role models.UserRole) (*models.BankingUser, error) {
	if svc.Store.DB != nil {
		var user models.BankingUser
		err := svc.Store.DB.Where("email = ?", email).First(&user).Error

		if err == nil {
			// User exists, update auth provider if needed
			if user.AuthProvider == "" || user.AuthProvider == "LOCAL" {
				user.AuthProvider = provider
				user.ProviderID = providerID
				svc.Store.DB.Save(&user)
			}
			return &user, nil
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}

		// New user registration via OAuth
		now := time.Now().UTC()
		kycStatus := "PENDING"
		kycTier := 0
		if role == models.RoleCustomer || role == models.RoleAdmin {
			kycStatus = "VERIFIED"
			kycTier = 1
		}

		newUser := &models.BankingUser{
			Name:          name,
			Email:         email,
			Role:          role,
			KYCStatus:     kycStatus,
			KYCTier:       kycTier,
			AuthProvider:  provider,
			ProviderID:    providerID,
			GDPRConsentAt: &now,
			Country:       "IT",
		}

		if err := svc.Store.DB.Create(newUser).Error; err != nil {
			return nil, err
		}

		// Synchronize with legacy in-memory user map
		svc.Store.Lock()
		svc.Store.Users[newUser.ID.String()] = &models.User{
			ID:        newUser.ID.String(),
			Name:      newUser.Name,
			Email:     newUser.Email,
			Role:      newUser.Role,
			KYCStatus: newUser.KYCStatus,
		}
		svc.Store.Unlock()

		return newUser, nil
	}

	// Memory fallback
	uid := uuid.New()
	newUser := &models.BankingUser{
		BaseModel:    models.BaseModel{ID: uid},
		Name:         name,
		Email:        email,
		Role:         role,
		KYCStatus:    "VERIFIED",
		AuthProvider: provider,
		ProviderID:   providerID,
	}
	svc.Store.Lock()
	svc.Store.Users[uid.String()] = &models.User{
		ID:        uid.String(),
		Name:      name,
		Email:     email,
		Role:      role,
		KYCStatus: "VERIFIED",
	}
	svc.Store.Unlock()
	return newUser, nil
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

// SubmitCustomerKYC advances a customer's KYC Tier under EU AMLD5 / MiCA guidelines
func (svc *Services) SubmitCustomerKYC(userID string, tier int, docType, docNumber string) (*models.BankingUser, error) {
	var user models.BankingUser
	if err := svc.Store.DB.Where("id = ? OR email = ?", userID, userID).First(&user).Error; err != nil {
		// Fallback for demo / memory users
		if legacyUser, ok := svc.Store.Users[userID]; ok {
			legacyUser.KYCStatus = "VERIFIED"
			return &models.BankingUser{
				Name:      legacyUser.Name,
				Email:     legacyUser.Email,
				Role:      legacyUser.Role,
				KYCStatus: "VERIFIED",
				KYCTier:   tier,
			}, nil
		}
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	user.KYCTier = tier
	user.KYCDocType = docType
	user.KYCDocNumber = docNumber
	user.KYCStatus = "VERIFIED"

	if err := svc.Store.DB.Save(&user).Error; err != nil {
		return nil, err
	}

	// Synchronize in-memory legacy mirror
	svc.Store.Lock()
	if memUser, ok := svc.Store.Users[user.ID.String()]; ok {
		memUser.KYCStatus = "VERIFIED"
	}
	svc.Store.Unlock()

	return &user, nil
}


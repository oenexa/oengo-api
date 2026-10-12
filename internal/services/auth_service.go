package services

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"oengo-api/internal/apperrors"
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
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" || !strings.Contains(req.Email, "@") {
		return nil, apperrors.NewValidation("email", "a valid email address is required")
	}
	if req.Name == "" {
		return nil, apperrors.NewValidation("name", "full name is required")
	}
	if len(req.Password) < 6 {
		return nil, apperrors.NewValidation("password", "password must be at least 6 characters")
	}

	// 1. Check if email exists
	if svc.Store.DB != nil {
		var existing models.BankingUser
		if err := svc.Store.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
			return nil, apperrors.NewConflict("user with email already exists")
		} else if err != gorm.ErrRecordNotFound {
			return nil, apperrors.WrapDBError("check_existing_user", err)
		}
	} else {
		svc.Store.RLock()
		for _, u := range svc.Store.Users {
			if strings.EqualFold(u.Email, req.Email) {
				svc.Store.RUnlock()
				return nil, apperrors.NewConflict("user with email already exists")
			}
		}
		svc.Store.RUnlock()
	}

	// 2. Hash Password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.NewInternal("failed to hash password", err)
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
			return nil, apperrors.WrapDBError("register_banking_user", err)
		}
	} else {
		user.ID = uuid.New()
	}

	// 4. Synchronize with store & relational tables
	svc.Store.Lock()
	defer svc.Store.Unlock()

	legacyUser := &models.User{
		ID:           user.ID.String(),
		Name:         user.Name,
		Email:        user.Email,
		Phone:        user.Phone,
		PasswordHash: user.PasswordHash,
		Role:         user.Role,
		KYCStatus:    user.KYCStatus,
	}
	if err := svc.Store.SaveUser(legacyUser); err != nil {
		return nil, err
	}

	if user.Role == models.RoleCourier {
		rider := &models.RiderProfile{
			ID:        user.ID.String(),
			Name:      user.Name,
			Phone:     user.Phone,
			KYCStatus: user.KYCStatus,
			IsOnline:  true,
		}
		if err := svc.Store.SaveRider(rider); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (svc *Services) LoginUser(email, password string) (*models.BankingUser, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return nil, apperrors.NewValidation("credentials", "email and password are required")
	}

	if svc.Store.DB != nil {
		var user models.BankingUser
		if err := svc.Store.DB.Where("email = ?", email).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, apperrors.NewUnauthorized("user not found")
			}
			return nil, apperrors.WrapDBError("login_lookup", err)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			return nil, apperrors.NewUnauthorized("invalid credentials")
		}

		return &user, nil
	}

	// In-memory fallback
	svc.Store.RLock()
	defer svc.Store.RUnlock()
	for _, u := range svc.Store.Users {
		if strings.EqualFold(u.Email, email) {
			if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
				return nil, apperrors.NewUnauthorized("invalid credentials")
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
	return nil, apperrors.NewUnauthorized("user not found")
}

// LoginOrRegisterOAuth supports seamless single-click Google/Gmail and Passkey authentication
func (svc *Services) LoginOrRegisterOAuth(name, email, provider, providerID string, role models.UserRole) (*models.BankingUser, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, apperrors.NewValidation("email", "valid OAuth email is required")
	}
	if name == "" {
		name = "Verified OENGO User"
	}

	if svc.Store.DB != nil {
		var user models.BankingUser
		err := svc.Store.DB.Where("email = ?", email).First(&user).Error

		if err == nil {
			if user.AuthProvider == "" || user.AuthProvider == "LOCAL" {
				user.AuthProvider = provider
				user.ProviderID = providerID
				if saveErr := svc.Store.DB.Save(&user).Error; saveErr != nil {
					return nil, apperrors.WrapDBError("oauth_update_provider", saveErr)
				}
			}
			return &user, nil
		} else if err != gorm.ErrRecordNotFound {
			return nil, apperrors.WrapDBError("oauth_user_lookup", err)
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
			return nil, apperrors.WrapDBError("oauth_create_user", err)
		}

		svc.Store.Lock()
		_ = svc.Store.SaveUser(&models.User{
			ID:        newUser.ID.String(),
			Name:      newUser.Name,
			Email:     newUser.Email,
			Role:      newUser.Role,
			KYCStatus: newUser.KYCStatus,
		})
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
	_ = svc.Store.SaveUser(&models.User{
		ID:        uid.String(),
		Name:      name,
		Email:     email,
		Role:      role,
		KYCStatus: "VERIFIED",
	})
	svc.Store.Unlock()
	return newUser, nil
}

func (svc *Services) ApproveKYC(userID string) (*models.User, error) {
	svc.Store.Lock()
	defer svc.Store.Unlock()

	u, exists := svc.Store.Users[userID]
	if !exists {
		return nil, apperrors.NewNotFound("user", userID)
	}

	u.KYCStatus = "VERIFIED"
	if err := svc.Store.SaveUser(u); err != nil {
		return nil, err
	}

	if u.Role == models.RoleCourier {
		if rp, ok := svc.Store.Riders[userID]; ok {
			rp.KYCStatus = "VERIFIED"
			if err := svc.Store.SaveRider(rp); err != nil {
				return nil, err
			}
		}
	}
	return u, nil
}

// SubmitCustomerKYC advances a customer's KYC Tier under EU AMLD5 / MiCA guidelines
func (svc *Services) SubmitCustomerKYC(userID string, tier int, docType, docNumber string) (*models.BankingUser, error) {
	if tier < 1 || tier > 3 {
		return nil, apperrors.NewValidation("tier", "KYC tier must be between 1 and 3")
	}

	if svc.Store.DB != nil {
		var user models.BankingUser
		if err := svc.Store.DB.Where("id = ? OR email = ?", userID, userID).First(&user).Error; err == nil {
			user.KYCTier = tier
			user.KYCDocType = docType
			user.KYCDocNumber = docNumber
			user.KYCStatus = "VERIFIED"

			if err := svc.Store.DB.Save(&user).Error; err != nil {
				return nil, apperrors.WrapDBError("submit_customer_kyc", err)
			}

			svc.Store.Lock()
			if memUser, ok := svc.Store.Users[user.ID.String()]; ok {
				memUser.KYCStatus = "VERIFIED"
				_ = svc.Store.SaveUser(memUser)
			}
			svc.Store.Unlock()

			return &user, nil
		}
	}

	// Fallback for demo / memory users
	svc.Store.Lock()
	defer svc.Store.Unlock()
	if legacyUser, ok := svc.Store.Users[userID]; ok {
		legacyUser.KYCStatus = "VERIFIED"
		_ = svc.Store.SaveUser(legacyUser)
		return &models.BankingUser{
			Name:      legacyUser.Name,
			Email:     legacyUser.Email,
			Role:      legacyUser.Role,
			KYCStatus: "VERIFIED",
			KYCTier:   tier,
		}, nil
	}
	return nil, apperrors.NewNotFound("user", userID)
}


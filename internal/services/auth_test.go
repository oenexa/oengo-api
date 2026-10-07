package services_test

import (
	"testing"

	"oengo-api/internal/config"
	"oengo-api/internal/models"
	"oengo-api/internal/services"
	"oengo-api/internal/store"
)

func TestAuthAndKYC(t *testing.T) {
	cfg := &config.Config{}
	s := store.NewStore(cfg)
	svc := services.NewServices(s, cfg)

	// Test Registration
	req := services.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Phone:    "1234567890",
		Password: "password123",
		Role:     models.RoleCourier,
	}

	user, err := svc.RegisterUser(req)
	if err != nil {
		t.Fatalf("Expected no error on registration, got %v", err)
	}
	if user.Email != req.Email {
		t.Fatalf("Expected email %s, got %s", req.Email, user.Email)
	}
	if user.KYCStatus != "PENDING" {
		t.Fatalf("Expected Courier KYC Status to be PENDING, got %s", user.KYCStatus)
	}

	// Test Login
	loginUser, err := svc.LoginUser("test@example.com", "password123")
	if err != nil {
		t.Fatalf("Expected successful login, got error %v", err)
	}
	if loginUser.ID != user.ID {
		t.Fatalf("Expected matched IDs")
	}

	// Test Failed Login
	_, err = svc.LoginUser("test@example.com", "wrongpass")
	if err == nil {
		t.Fatalf("Expected error on invalid password")
	}

	// Test KYC Approval
	approvedUser, err := svc.ApproveKYC(user.ID)
	if err != nil {
		t.Fatalf("Expected successful KYC approval, got %v", err)
	}
	if approvedUser.KYCStatus != "VERIFIED" {
		t.Fatalf("Expected KYC VERIFIED, got %s", approvedUser.KYCStatus)
	}

	// Verify Courier Profile is also verified
	rider, err := svc.GetRiderProfile(user.ID)
	if err != nil {
		t.Fatalf("Expected to find rider profile")
	}
	if rider.KYCStatus != "VERIFIED" {
		t.Fatalf("Expected Rider Profile KYC VERIFIED, got %s", rider.KYCStatus)
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"oengo-api/internal/config"
	"oengo-api/internal/handlers"
	"oengo-api/internal/models"
	"oengo-api/internal/services"
	"oengo-api/internal/store"
)

func setupTestApp() (*handlers.Router, *services.Services) {
	cfg := config.LoadConfig()
	st := store.NewStore(cfg)
	svc := services.NewServices(st, cfg)
	engine := handlers.SetupRouter(svc)
	return &handlers.Router{Engine: engine, Services: svc}, svc
}

func TestHealthCheck(t *testing.T) {
	router, _ := setupTestApp()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if resp["status"] != "UP" || resp["service"] != "oengo-api" {
		t.Fatalf("unexpected health payload: %v", resp)
	}
}

func TestListRestaurantsAndMenu(t *testing.T) {
	router, _ := setupTestApp()

	req, _ := http.NewRequest("GET", "/api/restaurants", nil)
	w := httptest.NewRecorder()
	router.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var res struct {
		Success     bool                 `json:"success"`
		Count       int                  `json:"count"`
		Restaurants []*models.Restaurant `json:"restaurants"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if res.Count < 4 {
		t.Fatalf("expected at least 4 restaurants, got %d", res.Count)
	}

	// Test Menu
	reqMenu, _ := http.NewRequest("GET", "/api/restaurants/user_restaurant/menu", nil)
	wMenu := httptest.NewRecorder()
	router.Engine.ServeHTTP(wMenu, reqMenu)

	if wMenu.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", wMenu.Code)
	}
}

func TestOrderLifecycleAndLedgerDisbursal(t *testing.T) {
	router, svc := setupTestApp()

	// 1. Create Order
	orderReq := services.CreateOrderRequest{
		BuyerID:       "user_customer",
		RestaurantID:  "user_restaurant",
		Amount:        20.00,
		DeliveryFee:   3.00,
		Tip:           2.00,
		PaymentMethod: "CREDIT_CARD",
		CardPayment: &models.CardPaymentInfo{
			Brand: "Visa",
			Last4: "4242",
		},
	}
	body, _ := json.Marshal(orderReq)

	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var orderResp struct {
		Order models.Order `json:"order"`
	}
	json.Unmarshal(w.Body.Bytes(), &orderResp)
	orderID := orderResp.Order.ID
	pin := orderResp.Order.DeliveryPIN

	if orderResp.Order.Status != models.StatusAwaitingRestaurant {
		t.Fatalf("expected status AWAITING_RESTAURANT, got %s", orderResp.Order.Status)
	}

	// 2. Accept Order
	reqAccept, _ := http.NewRequest("POST", "/api/orders/"+orderID+"/accept", nil)
	wAccept := httptest.NewRecorder()
	router.Engine.ServeHTTP(wAccept, reqAccept)
	if wAccept.Code != http.StatusOK {
		t.Fatalf("accept failed: %d", wAccept.Code)
	}

	// 3. Mark Ready
	reqReady, _ := http.NewRequest("POST", "/api/orders/"+orderID+"/ready", nil)
	wReady := httptest.NewRecorder()
	router.Engine.ServeHTTP(wReady, reqReady)
	if wReady.Code != http.StatusOK {
		t.Fatalf("ready failed: %d", wReady.Code)
	}

	// 4. Confirm Delivery with PIN
	delivBody, _ := json.Marshal(map[string]string{"code": pin})
	reqDeliv, _ := http.NewRequest("POST", "/api/orders/"+orderID+"/confirm-delivery", bytes.NewBuffer(delivBody))
	reqDeliv.Header.Set("Content-Type", "application/json")
	wDeliv := httptest.NewRecorder()
	router.Engine.ServeHTTP(wDeliv, reqDeliv)

	if wDeliv.Code != http.StatusOK {
		t.Fatalf("confirm delivery failed: %d: %s", wDeliv.Code, wDeliv.Body.String())
	}

	// 5. Verify Invariants: sum of debits == sum of credits
	ledger := svc.GetLedgerEntries(100)
	var totalDebits, totalCredits float64
	for _, entry := range ledger {
		if entry.EntryType == models.LedgerDebit {
			totalDebits += entry.AmountEUR
		} else if entry.EntryType == models.LedgerCredit {
			totalCredits += entry.AmountEUR
		}
	}

	if totalDebits != totalCredits {
		t.Fatalf("ledger unbalanced: debits %.2f != credits %.2f", totalDebits, totalCredits)
	}
}

func TestWalletAndCoinEngine(t *testing.T) {
	router, _ := setupTestApp()

	// Deposit
	depBody, _ := json.Marshal(map[string]interface{}{
		"userId": "user_customer",
		"amount": 25.00,
	})
	reqDep, _ := http.NewRequest("POST", "/api/wallet/deposit", bytes.NewBuffer(depBody))
	reqDep.Header.Set("Content-Type", "application/json")
	wDep := httptest.NewRecorder()
	router.Engine.ServeHTTP(wDep, reqDep)

	if wDep.Code != http.StatusOK {
		t.Fatalf("deposit failed: %d", wDep.Code)
	}

	// Coins Redeem
	redeemBody, _ := json.Marshal(map[string]interface{}{
		"userId":   "user_customer",
		"coins":    200,
		"subtotal": 20.00,
	})
	reqRedeem, _ := http.NewRequest("POST", "/api/coins/redeem", bytes.NewBuffer(redeemBody))
	reqRedeem.Header.Set("Content-Type", "application/json")
	wRedeem := httptest.NewRecorder()
	router.Engine.ServeHTTP(wRedeem, reqRedeem)

	if wRedeem.Code != http.StatusOK {
		t.Fatalf("redeem coins failed: %d: %s", wRedeem.Code, wRedeem.Body.String())
	}
}

func TestRiderAndAdminDashboard(t *testing.T) {
	router, _ := setupTestApp()

	// Rider Toggle Status
	rBody, _ := json.Marshal(map[string]interface{}{
		"riderId":  "user_courier",
		"isOnline": true,
	})
	reqRider, _ := http.NewRequest("POST", "/api/rider/status", bytes.NewBuffer(rBody))
	reqRider.Header.Set("Content-Type", "application/json")
	wRider := httptest.NewRecorder()
	router.Engine.ServeHTTP(wRider, reqRider)

	if wRider.Code != http.StatusOK {
		t.Fatalf("rider status failed: %d", wRider.Code)
	}

	// Admin Dashboard
	reqAdmin, _ := http.NewRequest("GET", "/api/admin/dashboard", nil)
	wAdmin := httptest.NewRecorder()
	router.Engine.ServeHTTP(wAdmin, reqAdmin)

	if wAdmin.Code != http.StatusOK {
		t.Fatalf("admin dashboard failed: %d", wAdmin.Code)
	}
}

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOenexaCoreBridge_AuthorizeCheckKycAndNotify(t *testing.T) {
	var seenIdempotencyKey string
	var seenAuthHeader string
	var seenSubHeader string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/payments/intents":
			seenIdempotencyKey = r.Header.Get("Idempotency-Key")
			seenAuthHeader = r.Header.Get("Authorization")
			seenSubHeader = r.Header.Get("X-Oenexa-Sub")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(PaymentIntentResponse{
				PaymentIntentID: "pi_core_123",
				Status:          "AUTHORIZED",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/kyc/status/usr-001":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(KycStatusResponse{
				SubjectUUID: "usr-001",
				Tier:        "TIER2",
				Status:      "APPROVED",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/notifications/dispatch":
			w.WriteHeader(http.StatusAccepted)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	bridge := NewOenexaCoreBridge(srv.URL, "svc-jwt-token")
	if !bridge.Enabled() {
		t.Fatal("expected bridge to be enabled")
	}

	pi, err := bridge.AuthorizeOrderPayment(context.Background(), PaymentIntentRequest{
		UserID:         "usr-001",
		OrderID:        "ord_55",
		Amount:         19.95,
		Currency:       "EUR",
		Rail:           "ONXC",
		IdempotencyKey: "idem-ord-55",
	}, "corr-xyz")
	if err != nil {
		t.Fatalf("AuthorizeOrderPayment failed: %v", err)
	}
	if pi.PaymentIntentID != "pi_core_123" || pi.Status != "AUTHORIZED" {
		t.Fatalf("unexpected payment intent response: %+v", pi)
	}
	if seenIdempotencyKey != "idem-ord-55" || seenAuthHeader != "Bearer svc-jwt-token" || seenSubHeader != "usr-001" {
		t.Fatalf("missing expected headers: idem=%q auth=%q sub=%q", seenIdempotencyKey, seenAuthHeader, seenSubHeader)
	}

	kyc, err := bridge.CheckCustomerKycTier(context.Background(), "usr-001", "corr-xyz")
	if err != nil {
		t.Fatalf("CheckCustomerKycTier failed: %v", err)
	}
	if kyc.Tier != "TIER2" || kyc.Status != "APPROVED" {
		t.Fatalf("unexpected kyc status: %+v", kyc)
	}

	if err := bridge.DispatchNotification(context.Background(), NotificationEvent{
		SubjectUserID: "usr-001",
		Subject:       "Order Confirmed",
		Body:          "Order ord_55 paid via ONXC",
	}, "corr-xyz"); err != nil {
		t.Fatalf("DispatchNotification failed: %v", err)
	}
}

func TestOenexaCoreBridge_DisabledReturnsClearError(t *testing.T) {
	bridge := NewOenexaCoreBridge("", "")
	if bridge.Enabled() {
		t.Fatal("expected bridge to report disabled when URL is empty")
	}
	if _, err := bridge.AuthorizeOrderPayment(context.Background(), PaymentIntentRequest{}, ""); err == nil {
		t.Fatal("expected error when bridge disabled")
	}
}

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// PaymentIntentRequest mirrors OENEXA Core /api/v1/payments/intents contract.
type PaymentIntentRequest struct {
	UserID         string  `json:"userId"`
	OrderID        string  `json:"orderId"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	Rail           string  `json:"rail"` // ONXC | WALLET | CARD
	IdempotencyKey string  `json:"idempotencyKey"`
}

// PaymentIntentResponse represents the authorized payment intent returned by OENEXA Pay.
type PaymentIntentResponse struct {
	PaymentIntentID string `json:"paymentIntentId"`
	Status          string `json:"status"`
}

// KycStatusResponse represents KYC tier verification metadata from oenexa-kyc-service.
type KycStatusResponse struct {
	SubjectUUID string `json:"subjectUuid"`
	Tier        string `json:"tier"`
	Status      string `json:"status"`
}

// NotificationEvent wraps canonical OENEXA notification.dispatch payloads.
type NotificationEvent struct {
	EventType     string            `json:"eventType"`
	SubjectUserID string            `json:"subjectUserId"`
	Email         string            `json:"email,omitempty"`
	Phone         string            `json:"phone,omitempty"`
	Subject       string            `json:"subject"`
	Body          string            `json:"body"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// OenexaCoreBridge connects OENGO commerce to api.oenexa.com Core services
// (oenexa-payment-service, oenexa-kyc-service, and oenexa-notification-service).
type OenexaCoreBridge struct {
	gatewayURL   string
	serviceToken string
	httpClient   *http.Client
}

// NewOenexaCoreBridge creates a bridge client targeting gatewayURL with an optional bearer token.
func NewOenexaCoreBridge(gatewayURL, serviceToken string) *OenexaCoreBridge {
	trimmed := strings.TrimRight(strings.TrimSpace(gatewayURL), "/")
	return &OenexaCoreBridge{
		gatewayURL:   trimmed,
		serviceToken: strings.TrimSpace(serviceToken),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// NewOenexaCoreBridgeFromEnv constructs the bridge from OENEXA_GATEWAY_URL and OENEXA_INTERNAL_SERVICE_TOKEN.
func NewOenexaCoreBridgeFromEnv() *OenexaCoreBridge {
	return NewOenexaCoreBridge(
		os.Getenv("OENEXA_GATEWAY_URL"),
		os.Getenv("OENEXA_INTERNAL_SERVICE_TOKEN"),
	)
}

// Enabled reports whether an upstream OENEXA Gateway URL is configured.
func (b *OenexaCoreBridge) Enabled() bool {
	return b != nil && b.gatewayURL != ""
}

// AuthorizeOrderPayment initiates a tri-rail payment intent via oenexa-payment-service.
func (b *OenexaCoreBridge) AuthorizeOrderPayment(ctx context.Context, req PaymentIntentRequest, correlationID string) (*PaymentIntentResponse, error) {
	if !b.Enabled() {
		return nil, fmt.Errorf("oenexa bridge disabled: OENEXA_GATEWAY_URL not configured")
	}
	if req.Currency == "" {
		req.Currency = "EUR"
	}
	if req.IdempotencyKey == "" {
		req.IdempotencyKey = fmt.Sprintf("oengo-%s-%d", req.OrderID, time.Now().UnixNano())
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, b.gatewayURL+"/api/v1/payments/intents", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	b.applyHeaders(httpReq, req.UserID, correlationID)
	httpReq.Header.Set("Idempotency-Key", req.IdempotencyKey)

	resp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("payment service returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var parsed PaymentIntentResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, err
	}
	return &parsed, nil
}

// CheckCustomerKycTier queries centralized KYC tier status from oenexa-kyc-service.
func (b *OenexaCoreBridge) CheckCustomerKycTier(ctx context.Context, subjectUUID, correlationID string) (*KycStatusResponse, error) {
	if !b.Enabled() {
		return nil, fmt.Errorf("oenexa bridge disabled: OENEXA_GATEWAY_URL not configured")
	}
	url := fmt.Sprintf("%s/api/v1/kyc/status/%s", b.gatewayURL, subjectUUID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	b.applyHeaders(httpReq, subjectUUID, correlationID)

	resp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("kyc service returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var parsed KycStatusResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, err
	}
	return &parsed, nil
}

// DispatchNotification forwards order/payment notification events to oenexa-notification-service.
func (b *OenexaCoreBridge) DispatchNotification(ctx context.Context, evt NotificationEvent, correlationID string) error {
	if !b.Enabled() {
		return fmt.Errorf("oenexa bridge disabled: OENEXA_GATEWAY_URL not configured")
	}
	if evt.EventType == "" {
		evt.EventType = "notification.dispatch"
	}
	payload, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, b.gatewayURL+"/api/v1/notifications/dispatch", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	b.applyHeaders(httpReq, evt.SubjectUserID, correlationID)

	resp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
		return fmt.Errorf("notification service returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}
	return nil
}

func (b *OenexaCoreBridge) applyHeaders(req *http.Request, subjectUUID, correlationID string) {
	req.Header.Set("Content-Type", "application/json")
	if b.serviceToken != "" {
		req.Header.Set("Authorization", "Bearer "+b.serviceToken)
	}
	if subjectUUID != "" {
		req.Header.Set("X-Oenexa-Sub", subjectUUID)
	}
	if correlationID != "" {
		req.Header.Set("X-Correlation-Id", correlationID)
	}
}

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// OrderStatus defines the strict state machine for food delivery escrow
type OrderStatus string

const (
	StatusCreated    OrderStatus = "CREATED"
	StatusPreparing  OrderStatus = "PREPARING"
	StatusInTransit  OrderStatus = "IN_TRANSIT"
	StatusDelivered  OrderStatus = "DELIVERED"
	StatusRefunded   OrderStatus = "REFUNDED"
	StatusDisputed   OrderStatus = "DISPUTED"
)

// EscrowOrder represents an on-chain food delivery order
type EscrowOrder struct {
	ID                string      `json:"id"`
	Buyer             string      `json:"buyer"`
	Restaurant        string      `json:"restaurant"`
	Courier           string      `json:"courier"`
	Amount            uint64      `json:"amount"`       // Food subtotal in OEN
	DeliveryFee       uint64      `json:"delivery_fee"` // Delivery fee in OEN
	Tip               uint64      `json:"tip"`          // Courier tip in OEN
	TotalLocked       uint64      `json:"total_locked"` // Total funds locked in contract
	Status            OrderStatus `json:"status"`
	PickupBarcodeHash string      `json:"pickup_barcode_hash"` // SHA256 of pickup barcode
	DeliveryProofHash string      `json:"delivery_proof_hash"` // SHA256 of delivery barcode/PIN
	CreatedAt         int64       `json:"created_at"`
	SettledAt         int64       `json:"settled_at,omitempty"`
	RestaurantPayout  uint64      `json:"restaurant_payout,omitempty"`
	CourierPayout     uint64      `json:"courier_payout,omitempty"`
}

var (
	// contractState persists order states across contract calls
	contractState = make(map[string]EscrowOrder)
)

// hashString computes SHA256 hex string for verification
func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// Call is the entry point for the WASM smart contract execution
func Call(method string, paramsJSON string) (string, error) {
	switch method {

	case "createOrder":
		var params struct {
			ID                string `json:"id"`
			Buyer             string `json:"buyer"`
			Restaurant        string `json:"restaurant"`
			Amount            uint64 `json:"amount"`
			DeliveryFee       uint64 `json:"delivery_fee"`
			Tip               uint64 `json:"tip"`
			PickupBarcode     string `json:"pickup_barcode"`
			DeliverySecretPIN string `json:"delivery_secret_pin"`
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return "", fmt.Errorf("invalid createOrder params: %w", err)
		}
		if params.ID == "" || params.Buyer == "" || params.Restaurant == "" {
			return "", fmt.Errorf("id, buyer, and restaurant are required")
		}
		if params.Amount == 0 {
			return "", fmt.Errorf("order amount must be greater than zero")
		}

		total := params.Amount + params.DeliveryFee + params.Tip
		order := EscrowOrder{
			ID:                params.ID,
			Buyer:             params.Buyer,
			Restaurant:        params.Restaurant,
			Amount:            params.Amount,
			DeliveryFee:       params.DeliveryFee,
			Tip:               params.Tip,
			TotalLocked:       total,
			Status:            StatusCreated,
			PickupBarcodeHash: hashString(params.PickupBarcode),
			DeliveryProofHash: hashString(params.DeliverySecretPIN),
			CreatedAt:         time.Now().Unix(),
		}
		contractState[params.ID] = order

		return fmt.Sprintf("Order %s created and %d OEN locked in escrow", params.ID, total), nil

	case "acceptOrder":
		var params struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return "", err
		}
		order, exists := contractState[params.ID]
		if !exists {
			return "", fmt.Errorf("order %s not found", params.ID)
		}
		if order.Status != StatusCreated {
			return "", fmt.Errorf("cannot accept order with status %s", order.Status)
		}

		order.Status = StatusPreparing
		contractState[params.ID] = order
		return fmt.Sprintf("Order %s accepted by restaurant, now preparing", params.ID), nil

	case "assignCourier":
		var params struct {
			ID      string `json:"id"`
			Courier string `json:"courier"`
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return "", err
		}
		order, exists := contractState[params.ID]
		if !exists {
			return "", fmt.Errorf("order %s not found", params.ID)
		}
		if order.Status != StatusPreparing && order.Status != StatusCreated {
			return "", fmt.Errorf("cannot assign courier to order with status %s", order.Status)
		}

		order.Courier = params.Courier
		contractState[params.ID] = order
		return fmt.Sprintf("Courier %s assigned to order %s", params.Courier, params.ID), nil

	case "confirmPickup":
		var params struct {
			ID            string `json:"id"`
			PickupBarcode string `json:"pickup_barcode"`
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return "", err
		}
		order, exists := contractState[params.ID]
		if !exists {
			return "", fmt.Errorf("order %s not found", params.ID)
		}
		if order.Status != StatusPreparing {
			return "", fmt.Errorf("order is not ready for pickup (current status: %s)", order.Status)
		}
		if order.Courier == "" {
			return "", fmt.Errorf("no courier assigned to order %s", params.ID)
		}
		if hashString(params.PickupBarcode) != order.PickupBarcodeHash {
			return "", fmt.Errorf("invalid pickup barcode verification")
		}

		order.Status = StatusInTransit
		contractState[params.ID] = order
		return fmt.Sprintf("Pickup confirmed for order %s. Order is now IN_TRANSIT", params.ID), nil

	case "confirmDelivery":
		var params struct {
			ID                string `json:"id"`
			DeliveryProofCode string `json:"delivery_proof_code"` // Customer Barcode or PIN
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return "", err
		}
		order, exists := contractState[params.ID]
		if !exists {
			return "", fmt.Errorf("order %s not found", params.ID)
		}
		if order.Status != StatusInTransit {
			return "", fmt.Errorf("order must be IN_TRANSIT to confirm delivery (current: %s)", order.Status)
		}
		if hashString(params.DeliveryProofCode) != order.DeliveryProofHash {
			return "", fmt.Errorf("invalid delivery proof: barcode or PIN does not match customer secret")
		}

		// Autonomous Settlement:
		// 95% of food subtotal to Restaurant
		// 5% of food subtotal + 100% of DeliveryFee + 100% of Tip to Courier
		restaurantCut := (order.Amount * 95) / 100
		courierCut := (order.Amount - restaurantCut) + order.DeliveryFee + order.Tip

		order.Status = StatusDelivered
		order.RestaurantPayout = restaurantCut
		order.CourierPayout = courierCut
		order.SettledAt = time.Now().Unix()
		contractState[params.ID] = order

		return fmt.Sprintf("Order %s delivered! Escrow settled: %d OEN to Restaurant (%s), %d OEN to Courier (%s)",
			params.ID, restaurantCut, order.Restaurant, courierCut, order.Courier), nil

	case "refundOrder":
		var params struct {
			ID     string `json:"id"`
			Reason string `json:"reason"`
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return "", err
		}
		order, exists := contractState[params.ID]
		if !exists {
			return "", fmt.Errorf("order %s not found", params.ID)
		}
		if order.Status == StatusDelivered {
			return "", fmt.Errorf("cannot refund already delivered order")
		}

		order.Status = StatusRefunded
		order.SettledAt = time.Now().Unix()
		contractState[params.ID] = order

		return fmt.Sprintf("Order %s refunded 100%% (%d OEN) to Buyer (%s). Reason: %s",
			params.ID, order.TotalLocked, order.Buyer, params.Reason), nil

	case "getOrder":
		var params struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return "", err
		}
		order, exists := contractState[params.ID]
		if !exists {
			return "", fmt.Errorf("order %s not found", params.ID)
		}
		res, err := json.Marshal(order)
		if err != nil {
			return "", err
		}
		return string(res), nil

	default:
		return "", fmt.Errorf("unknown smart contract method: %s", method)
	}
}

func main() {
	// Entry point for Go WASM compilation
}

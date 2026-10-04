package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEscrowCompleteLifecycle(t *testing.T) {
	orderID := "ord_test_001"
	buyer := "0xBuyerAddress_MLDSA"
	restaurant := "0xRestaurantAddress_MLDSA"
	courier := "0xCourierAddress_MLDSA"
	amount := uint64(1000)
	fee := uint64(100)
	tip := uint64(50)
	pickupBarcode := "PKG-BARCODE-999"
	deliveryPIN := "4821"

	// 1. Create Order
	createParams, _ := json.Marshal(map[string]interface{}{
		"id":                  orderID,
		"buyer":               buyer,
		"restaurant":          restaurant,
		"amount":              amount,
		"delivery_fee":        fee,
		"tip":                 tip,
		"pickup_barcode":      pickupBarcode,
		"delivery_secret_pin": deliveryPIN,
	})
	res, err := Call("createOrder", string(createParams))
	if err != nil {
		t.Fatalf("createOrder failed: %v", err)
	}
	if !strings.Contains(res, "1150 OEN locked") {
		t.Errorf("expected 1150 total locked, got %s", res)
	}

	// 2. Accept Order by Restaurant
	acceptParams, _ := json.Marshal(map[string]string{"id": orderID})
	res, err = Call("acceptOrder", string(acceptParams))
	if err != nil {
		t.Fatalf("acceptOrder failed: %v", err)
	}
	if !strings.Contains(res, "now preparing") {
		t.Errorf("unexpected acceptOrder response: %s", res)
	}

	// 3. Assign Courier
	assignParams, _ := json.Marshal(map[string]string{
		"id":      orderID,
		"courier": courier,
	})
	res, err = Call("assignCourier", string(assignParams))
	if err != nil {
		t.Fatalf("assignCourier failed: %v", err)
	}
	if !strings.Contains(res, "Courier 0xCourierAddress_MLDSA assigned") {
		t.Errorf("unexpected assignCourier response: %s", res)
	}

	// 4. Test Invalid Pickup Barcode Rejection
	badPickupParams, _ := json.Marshal(map[string]string{
		"id":             orderID,
		"pickup_barcode": "WRONG-BARCODE",
	})
	_, err = Call("confirmPickup", string(badPickupParams))
	if err == nil {
		t.Fatalf("expected error on wrong pickup barcode, got nil")
	}

	// 5. Valid Pickup Barcode Scan
	goodPickupParams, _ := json.Marshal(map[string]string{
		"id":             orderID,
		"pickup_barcode": pickupBarcode,
	})
	res, err = Call("confirmPickup", string(goodPickupParams))
	if err != nil {
		t.Fatalf("confirmPickup failed: %v", err)
	}
	if !strings.Contains(res, "IN_TRANSIT") {
		t.Errorf("unexpected confirmPickup response: %s", res)
	}

	// 6. Test Invalid Delivery PIN Rejection
	badDeliveryParams, _ := json.Marshal(map[string]string{
		"id":                  orderID,
		"delivery_proof_code": "0000",
	})
	_, err = Call("confirmDelivery", string(badDeliveryParams))
	if err == nil {
		t.Fatalf("expected error on wrong delivery PIN, got nil")
	}

	// 7. Valid Delivery Handover Verification (PIN: 4821)
	goodDeliveryParams, _ := json.Marshal(map[string]string{
		"id":                  orderID,
		"delivery_proof_code": deliveryPIN,
	})
	res, err = Call("confirmDelivery", string(goodDeliveryParams))
	if err != nil {
		t.Fatalf("confirmDelivery failed: %v", err)
	}
	// 95% of 1000 = 950 to Restaurant.
	// (1000 - 950) + 100 fee + 50 tip = 200 to Courier.
	if !strings.Contains(res, "950 OEN to Restaurant") {
		t.Errorf("expected 950 to Restaurant, got %s", res)
	}
	if !strings.Contains(res, "200 OEN to Courier") {
		t.Errorf("expected 200 to Courier, got %s", res)
	}

	// 8. Query Order State
	getParams, _ := json.Marshal(map[string]string{"id": orderID})
	res, err = Call("getOrder", string(getParams))
	if err != nil {
		t.Fatalf("getOrder failed: %v", err)
	}
	var order EscrowOrder
	if err := json.Unmarshal([]byte(res), &order); err != nil {
		t.Fatalf("failed to parse getOrder JSON: %v", err)
	}
	if order.Status != StatusDelivered {
		t.Errorf("expected status DELIVERED, got %s", order.Status)
	}
	if order.RestaurantPayout != 950 || order.CourierPayout != 200 {
		t.Errorf("incorrect payouts: rest=%d, courier=%d", order.RestaurantPayout, order.CourierPayout)
	}
}

func TestEscrowRefundFlow(t *testing.T) {
	orderID := "ord_refund_002"
	buyer := "0xBuyerRefund"
	restaurant := "0xRestaurantRefund"

	// Create Order
	createParams, _ := json.Marshal(map[string]interface{}{
		"id":                  orderID,
		"buyer":               buyer,
		"restaurant":          restaurant,
		"amount":              uint64(500),
		"delivery_fee":        uint64(50),
		"tip":                 uint64(0),
		"pickup_barcode":      "BAR123",
		"delivery_secret_pin": "1234",
	})
	_, err := Call("createOrder", string(createParams))
	if err != nil {
		t.Fatalf("createOrder failed: %v", err)
	}

	// Refund Order
	refundParams, _ := json.Marshal(map[string]string{
		"id":     orderID,
		"reason": "Restaurant out of stock",
	})
	res, err := Call("refundOrder", string(refundParams))
	if err != nil {
		t.Fatalf("refundOrder failed: %v", err)
	}
	if !strings.Contains(res, "refunded 100% (550 OEN)") {
		t.Errorf("unexpected refund response: %s", res)
	}
}

func TestEscrowConfigurableCommission(t *testing.T) {
	// 1. Test Rejection of Excessive Commission (> 30%)
	badRateParams, _ := json.Marshal(map[string]uint64{"rate_pct": 35})
	_, err := Call("setCommissionRate", string(badRateParams))
	if err == nil {
		t.Fatalf("expected error when setting commission rate > 30%%, got nil")
	}

	// 2. Test Setting Valid Global Commission Rate (e.g. 10%)
	goodRateParams, _ := json.Marshal(map[string]uint64{"rate_pct": 10})
	res, err := Call("setCommissionRate", string(goodRateParams))
	if err != nil {
		t.Fatalf("setCommissionRate failed: %v", err)
	}
	if !strings.Contains(res, "updated to 10%") {
		t.Errorf("unexpected setCommissionRate response: %s", res)
	}

	// 3. Query getCommissionRate
	res, err = Call("getCommissionRate", "{}")
	if err != nil {
		t.Fatalf("getCommissionRate failed: %v", err)
	}
	if !strings.Contains(res, `"commission_rate_pct":10`) {
		t.Errorf("expected 10%% commission in getCommissionRate, got %s", res)
	}

	// 4. Create Order using Global 10% Commission
	orderID := "ord_comm_10"
	createParams, _ := json.Marshal(map[string]interface{}{
		"id":                  orderID,
		"buyer":               "0xBuyer10",
		"restaurant":          "0xRest10",
		"amount":              uint64(1000),
		"delivery_fee":        uint64(100),
		"tip":                 uint64(50),
		"pickup_barcode":      "PICK-10",
		"delivery_secret_pin": "5555",
	})
	res, err = Call("createOrder", string(createParams))
	if err != nil {
		t.Fatalf("createOrder failed: %v", err)
	}
	if !strings.Contains(res, "Commission: 10%") {
		t.Errorf("expected 10%% commission logged on createOrder, got %s", res)
	}

	// Move order to IN_TRANSIT
	acceptParams, _ := json.Marshal(map[string]string{"id": orderID})
	Call("acceptOrder", string(acceptParams))
	assignParams, _ := json.Marshal(map[string]string{"id": orderID, "courier": "0xCourier10"})
	Call("assignCourier", string(assignParams))
	pickupParams, _ := json.Marshal(map[string]string{"id": orderID, "pickup_barcode": "PICK-10"})
	Call("confirmPickup", string(pickupParams))

	// Confirm Delivery: 10% commission on 1000 -> 900 to Restaurant, (100 + 100 + 50) = 250 to Courier
	deliverParams, _ := json.Marshal(map[string]string{"id": orderID, "delivery_proof_code": "5555"})
	res, err = Call("confirmDelivery", string(deliverParams))
	if err != nil {
		t.Fatalf("confirmDelivery failed: %v", err)
	}
	if !strings.Contains(res, "900 OEN to Restaurant") {
		t.Errorf("expected 900 OEN to Restaurant, got %s", res)
	}
	if !strings.Contains(res, "250 OEN to Courier") {
		t.Errorf("expected 250 OEN to Courier, got %s", res)
	}

	// 5. Test 0% Promotional Commission Override on specific order
	zeroOrderID := "ord_comm_zero"
	zeroParams, _ := json.Marshal(map[string]interface{}{
		"id":                  zeroOrderID,
		"buyer":               "0xBuyerZero",
		"restaurant":          "0xRestZero",
		"amount":              uint64(1000),
		"delivery_fee":        uint64(100),
		"tip":                 uint64(50),
		"commission_rate_pct": uint64(0),
		"pickup_barcode":      "PICK-ZERO",
		"delivery_secret_pin": "0000",
	})
	res, err = Call("createOrder", string(zeroParams))
	if err != nil {
		t.Fatalf("createOrder with 0%% commission failed: %v", err)
	}
	if !strings.Contains(res, "Commission: 0%") {
		t.Errorf("expected 0%% commission logged on createOrder, got %s", res)
	}

	// Advance to Delivery
	acceptZero, _ := json.Marshal(map[string]string{"id": zeroOrderID})
	Call("acceptOrder", string(acceptZero))
	assignZero, _ := json.Marshal(map[string]string{"id": zeroOrderID, "courier": "0xCourierZero"})
	Call("assignCourier", string(assignZero))
	pickupZero, _ := json.Marshal(map[string]string{"id": zeroOrderID, "pickup_barcode": "PICK-ZERO"})
	Call("confirmPickup", string(pickupZero))

	// Confirm Delivery: 0% commission on 1000 -> 1000 to Restaurant (100%), (0 + 100 + 50) = 150 to Courier
	deliverZero, _ := json.Marshal(map[string]string{"id": zeroOrderID, "delivery_proof_code": "0000"})
	res, err = Call("confirmDelivery", string(deliverZero))
	if err != nil {
		t.Fatalf("confirmDelivery with 0%% commission failed: %v", err)
	}
	if !strings.Contains(res, "1000 OEN to Restaurant") {
		t.Errorf("expected 1000 OEN (100%%) to Restaurant, got %s", res)
	}
	if !strings.Contains(res, "150 OEN to Courier") {
		t.Errorf("expected 150 OEN to Courier, got %s", res)
	}

	// Reset default commission back to 5% for clean state
	resetParams, _ := json.Marshal(map[string]uint64{"rate_pct": 5})
	Call("setCommissionRate", string(resetParams))
}

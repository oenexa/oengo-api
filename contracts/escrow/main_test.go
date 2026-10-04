package main

import (
	"strings"
	"testing"
)

func TestEscrowCall(t *testing.T) {
	// 1. Test createOrder
	createParams := `{"id":"ord_101","buyer":"0xBuyerAddr","seller":"0xRestaurantAddr","amount":500}`
	res, err := Call("createOrder", createParams)
	if err != nil {
		t.Fatalf("createOrder failed: %v", err)
	}
	if !strings.Contains(res, "ord_101 created") {
		t.Errorf("unexpected createOrder response: %s", res)
	}

	// 2. Test confirmDelivery with existing order
	confirmParams := `{"id":"ord_101"}`
	res, err = Call("confirmDelivery", confirmParams)
	if err != nil {
		t.Fatalf("confirmDelivery failed: %v", err)
	}
	if !strings.Contains(res, "delivered, funds released") {
		t.Errorf("unexpected confirmDelivery response: %s", res)
	}

	// 3. Test unknown method
	_, err = Call("nonExistentMethod", `{}`)
	if err == nil {
		t.Fatalf("expected error on unknown method, got nil")
	}
}

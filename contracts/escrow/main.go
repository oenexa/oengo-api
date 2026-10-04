package main

import (
	"encoding/json"
	"fmt"
)

// DCommerceEscrow represents the smart contract state
type DCommerceEscrow struct {
	State map[string]Order `json:"state"`
}

// Order represents an e-commerce order
type Order struct {
	ID        string `json:"id"`
	Buyer     string `json:"buyer"`
	Seller    string `json:"seller"`
	Amount    uint64 `json:"amount"`
	Status    string `json:"status"` // PENDING, SHIPPED, DELIVERED, DISPUTED, REFUNDED
}

var (
	// contractState persists order states across contract calls
	contractState = make(map[string]Order)
)

// Call is the entry point for the WASM smart contract
func Call(method string, paramsJSON string) (string, error) {
	switch method {
	case "createOrder":
		var params struct {
			ID     string `json:"id"`
			Buyer  string `json:"buyer"`
			Seller string `json:"seller"`
			Amount uint64 `json:"amount"`
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return "", err
		}

		contractState[params.ID] = Order{
			ID:     params.ID,
			Buyer:  params.Buyer,
			Seller: params.Seller,
			Amount: params.Amount,
			Status: "PENDING",
		}
		// Funds would be locked here by calling the host environment
		return fmt.Sprintf("Order %s created and funds locked", params.ID), nil

	case "confirmDelivery":
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

		order.Status = "DELIVERED"
		contractState[params.ID] = order
		// Funds would be released to the seller here by calling the host environment
		return fmt.Sprintf("Order %s delivered, funds released to %s", params.ID, order.Seller), nil

	default:
		return "", fmt.Errorf("unknown method: %s", method)
	}
}

func main() {
	// The main function is required for Go WASM compilation, but execution begins in Call()
}

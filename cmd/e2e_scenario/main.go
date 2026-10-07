package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080/api"

func post(endpoint string, payload interface{}) (map[string]interface{}, error) {
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+endpoint, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %s", string(body))
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %v", resp.StatusCode, result["error"])
	}

	return result, nil
}

func get(endpoint string) (map[string]interface{}, error) {
	resp, err := http.Get(baseURL + endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func main() {
	fmt.Println("🚀 Starting Banking-Grade E2E Scenario...")

	// 1. Register Customer
	fmt.Println("\n[1] Registering Customer (Alice)...")
	customerRes, err := post("/auth/register", map[string]interface{}{
		"name":     "Alice Crypto",
		"email":    "alice@secure.com",
		"phone":    "+39000111222",
		"password": "BankGradePassword123!",
		"role":     "CUSTOMER",
	})
	if err != nil {
		panic(err)
	}
	customerID := customerRes["user"].(map[string]interface{})["id"].(string)
	fmt.Printf("✅ Customer Registered: %s\n", customerID)

	// 2. Register Restaurant
	fmt.Println("\n[2] Registering Restaurant (Gourmet Burgers)...")
	restRes, err := post("/auth/register", map[string]interface{}{
		"name":     "Gourmet Burgers Secure",
		"email":    "chef@burgers.com",
		"phone":    "+39999888777",
		"password": "BankGradePassword123!",
		"role":     "RESTAURANT",
	})
	if err != nil {
		panic(err)
	}
	restID := restRes["user"].(map[string]interface{})["id"].(string)
	fmt.Printf("✅ Restaurant Registered: %s\n", restID)

	// 3. Register Courier
	fmt.Println("\n[3] Registering Courier (Bob Speed)...")
	courierRes, err := post("/auth/register", map[string]interface{}{
		"name":     "Bob Speed",
		"email":    "bob@delivery.com",
		"phone":    "+39333444555",
		"password": "BankGradePassword123!",
		"role":     "COURIER",
	})
	if err != nil {
		panic(err)
	}
	courierID := courierRes["user"].(map[string]interface{})["id"].(string)
	fmt.Printf("✅ Courier Registered: %s\n", courierID)

	// 4. Admin Approves KYC
	fmt.Println("\n[4] Admin Approving KYC for Restaurant and Courier...")
	_, err = post("/admin/kyc/"+restID+"/approve", nil)
	if err != nil {
		panic(err)
	}
	_, err = post("/admin/kyc/"+courierID+"/approve", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("✅ KYC Approved securely.")

	// 5. Create Menu Item
	fmt.Println("\n[5] Restaurant Creating Menu Item...")
	menuRes, err := post("/restaurants/"+restID+"/menu", map[string]interface{}{
		"category":    "Burgers",
		"name":        "Crypto Double Smash Burger",
		"description": "Securely smashed double beef patty with zero-knowledge cheese.",
		"priceEUR":    15.50,
	})
	if err != nil {
		panic(err)
	}
	menuItem := menuRes["item"].(map[string]interface{})
	fmt.Printf("✅ Menu Item Created: %s (€%.2f)\n", menuItem["name"], menuItem["priceEUR"])

	// 6. Deposit Funds to Customer Wallet (Bank Transfer simulation)
	fmt.Println("\n[6] Depositing €50 to Customer Wallet...")
	_, err = post("/wallet/deposit", map[string]interface{}{
		"userId": customerID,
		"amount": 50.00,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("✅ Deposit successful. Ledger updated.")

	// 7. Place Order
	fmt.Println("\n[7] Customer Placing Live Order...")
	orderRes, err := post("/orders", map[string]interface{}{
		"buyerId":      customerID,
		"restaurantId": restID,
		"amount":       15.50,
		"deliveryFee":  2.50,
		"tip":          2.00,
		"paymentMethod": "WALLET",
		"items": []map[string]interface{}{
			{
				"id":    menuItem["id"],
				"name":  menuItem["name"],
				"price": menuItem["priceEUR"],
				"qty":   1,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	orderID := orderRes["order"].(map[string]interface{})["id"].(string)
	fmt.Printf("✅ Order placed and locked in Escrow! Order ID: %s\n", orderID)

	// 8. Order Lifecycle
	fmt.Println("\n[8] Executing Order Lifecycle (Accept -> Assign -> Pickup -> Delivery)...")
	_, err = post("/orders/"+orderID+"/accept", nil)
	if err != nil {
		panic(err)
	}
	
	_, err = post("/orders/"+orderID+"/ready", nil)
	if err != nil {
		panic(err)
	}

	_, err = post("/orders/"+orderID+"/assign-courier", map[string]interface{}{
		"courierId": courierID,
	})
	if err != nil {
		panic(err)
	}

	orderData, _ := get("/orders/" + orderID)
	orderObj := orderData["order"].(map[string]interface{})
	pickupBarcode := orderObj["pickupBarcode"].(string)
	deliveryPin := orderObj["deliveryPin"].(string)

	_, err = post("/orders/"+orderID+"/confirm-pickup", map[string]interface{}{
		"barcode": pickupBarcode,
	})
	if err != nil {
		panic(err)
	}

	_, err = post("/orders/"+orderID+"/confirm-delivery", map[string]interface{}{
		"code": deliveryPin,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("✅ Order Delivered Successfully! Escrow funds disbursed via Double-Entry Ledger.")

	// 9. Fetch Final Ledger State
	fmt.Println("\n[9] Fetching Banking Ledger State...")
	ledger, _ := get("/wallet/ledger?limit=4")
	entries := ledger["entries"].([]interface{})
	for _, e := range entries {
		entry := e.(map[string]interface{})
		fmt.Printf("   Ledger Entry: %s | Account: %s | %s | €%.2f\n", entry["id"], entry["accountCode"], entry["entryType"], entry["amountEUR"])
	}

	fmt.Println("\n🎉 END-TO-END SCENARIO COMPLETED SECURELY!")
}

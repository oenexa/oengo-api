package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"oengo-api/internal/models"
	"oengo-api/internal/services"
)

type Router struct {
	Engine   *gin.Engine
	Services *services.Services
}

func SetupRouter(svc *services.Services) *gin.Engine {
	r := gin.Default()

	// CORS Setup
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	r.Use(cors.New(config))

	// Health check (supports GET and HEAD)
	r.Match([]string{"GET", "HEAD"}, "/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "UP",
			"service":   "oengo-api",
			"runtime":   "golang-gin",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	api := r.Group("/api")
	{
		// ── RESTAURANTS & MENUS ──
		api.GET("/restaurants", func(c *gin.Context) {
			cuisine := c.Query("cuisine")
			search := c.Query("search")
			list := svc.ListRestaurants(cuisine, search)
			c.JSON(http.StatusOK, gin.H{
				"success":     true,
				"count":       len(list),
				"restaurants": list,
			})
		})

		api.GET("/restaurants/:id", func(c *gin.Context) {
			id := c.Param("id")
			r, err := svc.GetRestaurantProfile(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "restaurant": r})
		})

		api.PUT("/restaurants/:id", func(c *gin.Context) {
			id := c.Param("id")
			var body struct {
				IsOpen         *bool   `json:"isOpen"`
				PrepEtaMinutes *int    `json:"prepEtaMinutes"`
				Name           *string `json:"name"`
				Tagline        *string `json:"tagline"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			updated, err := svc.UpdateRestaurant(id, body.IsOpen, body.PrepEtaMinutes, body.Name, body.Tagline)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "restaurant": updated})
		})

		api.GET("/restaurants/:id/menu", func(c *gin.Context) {
			id := c.Param("id")
			menu := svc.GetMenu(id)
			c.JSON(http.StatusOK, gin.H{
				"success":      true,
				"restaurantId": id,
				"count":        len(menu),
				"menu":         menu,
			})
		})

		api.POST("/restaurants/:id/menu", func(c *gin.Context) {
			id := c.Param("id")
			var item models.MenuItem
			if err := c.ShouldBindJSON(&item); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			created, err := svc.AddDish(id, item)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"success": true, "item": created})
		})

		api.PUT("/restaurants/:id/menu/:itemId", func(c *gin.Context) {
			restID := c.Param("id")
			itemID := c.Param("itemId")
			var body struct {
				InStock       *bool    `json:"inStock"`
				PriceEUR      *float64 `json:"priceEUR"`
				StockQuantity *int     `json:"stockQuantity"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			updated, err := svc.UpdateDish(restID, itemID, body.InStock, body.PriceEUR, body.StockQuantity)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "item": updated})
		})

		api.DELETE("/restaurants/:id/menu/:itemId", func(c *gin.Context) {
			restID := c.Param("id")
			itemID := c.Param("itemId")
			if err := svc.DeleteDish(restID, itemID); err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Dish deleted"})
		})

		// ── ORDERS ──
		api.GET("/orders", func(c *gin.Context) {
			restID := c.Query("restaurantId")
			buyerID := c.Query("buyerId")
			status := c.Query("status")
			orders := svc.ListOrders(restID, buyerID, status)
			c.JSON(http.StatusOK, gin.H{"success": true, "count": len(orders), "orders": orders})
		})

		api.POST("/orders", func(c *gin.Context) {
			var req services.CreateOrderRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			order, err := svc.CreateOrder(req)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"message": "Order created and locked in escrow",
				"order":   order,
			})
		})

		api.GET("/orders/:id", func(c *gin.Context) {
			id := c.Param("id")
			order, err := svc.GetOrder(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "order": order})
		})

		api.GET("/orders/:id/track", func(c *gin.Context) {
			id := c.Param("id")
			order, err := svc.GetOrder(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			// Generate live tracking telemetry
			stage := 1
			progressPct := 15
			switch order.Status {
			case models.StatusPreparing:
				stage = 2
				progressPct = 40
			case models.StatusReadyForPickup:
				stage = 3
				progressPct = 65
			case models.StatusInTransit:
				stage = 4
				progressPct = 85
			case models.StatusDelivered:
				stage = 5
				progressPct = 100
			}

			c.JSON(http.StatusOK, gin.H{
				"success":       true,
				"orderId":       order.ID,
				"status":        order.Status,
				"stage":         stage,
				"progressPct":   progressPct,
				"pickupBarcode": order.PickupBarcode,
				"deliveryPin":   order.DeliveryPIN,
				"courier": gin.H{
					"name":        "Bob Delivery Rider",
					"vehicleType": "Electric Bicycle",
					"rating":      4.98,
					"lat":         40.8385,
					"lng":         14.2505,
				},
				"etaMinutes": 12,
			})
		})

		api.POST("/orders/:id/accept", func(c *gin.Context) {
			id := c.Param("id")
			updated, err := svc.UpdateOrderStatus(id, models.StatusPreparing)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "order": updated})
		})

		api.POST("/orders/:id/ready", func(c *gin.Context) {
			id := c.Param("id")
			updated, err := svc.UpdateOrderStatus(id, models.StatusReadyForPickup)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "order": updated})
		})

		api.POST("/orders/:id/assign-courier", func(c *gin.Context) {
			id := c.Param("id")
			var body struct {
				CourierID string `json:"courierId"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if body.CourierID == "" {
				body.CourierID = "user_courier"
			}
			updated, err := svc.AssignCourier(id, body.CourierID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "order": updated})
		})

		api.POST("/orders/:id/confirm-pickup", func(c *gin.Context) {
			id := c.Param("id")
			var body struct {
				Barcode string `json:"barcode"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			updated, err := svc.ConfirmPickup(id, body.Barcode)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "order": updated})
		})

		api.POST("/orders/:id/confirm-delivery", func(c *gin.Context) {
			id := c.Param("id")
			var body struct {
				Code string `json:"code"` // Pin or Barcode
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			updated, err := svc.ConfirmDelivery(id, body.Code)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success":          true,
				"message":          "Delivery verified and escrow disbursed",
				"order":            updated,
				"restaurantPayout": updated.RestaurantPayout,
				"courierPayout":    updated.CourierPayout,
			})
		})

		api.POST("/orders/:id/decline", func(c *gin.Context) {
			id := c.Param("id")
			var body struct {
				Reason string `json:"reason"`
			}
			c.ShouldBindJSON(&body)
			updated, err := svc.DeclineOrder(id, body.Reason)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "order": updated})
		})

		// ── PAYMENTS ──
		api.POST("/payments/card-intent", func(c *gin.Context) {
			var body struct {
				Amount   float64 `json:"amount"`
				Currency string  `json:"currency"`
			}
			c.ShouldBindJSON(&body)
			if body.Currency == "" {
				body.Currency = "EUR"
			}
			intent, err := svc.CreateCardIntent(body.Amount, body.Currency)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "intent": intent})
		})

		api.POST("/payments/confirm-card", func(c *gin.Context) {
			var body struct {
				PaymentIntentID string `json:"paymentIntentId"`
				CardNumber      string `json:"cardNumber"`
				CardHolderName  string `json:"cardHolderName"`
				SaveCard        bool   `json:"saveCard"`
				CustomerID      string `json:"customerId"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if body.CustomerID == "" {
				body.CustomerID = "user_customer"
			}
			cardInfo, err := svc.ConfirmCardPayment(body.PaymentIntentID, body.CardNumber, body.CardHolderName, body.SaveCard, body.CustomerID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success":       true,
				"message":       fmt.Sprintf("Payment authorized via %s", cardInfo.Brand),
				"cardPayment":   cardInfo,
				"status":        "SUCCEEDED",
				"transactionId": cardInfo.TransactionID,
			})
		})

		api.POST("/payments/instant-pay", func(c *gin.Context) {
			var body struct {
				UserID string  `json:"userId"`
				Method string  `json:"method"`
				Amount float64 `json:"amount"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if body.UserID == "" {
				body.UserID = "user_customer"
			}
			res, err := svc.InstantPay(body.UserID, body.Method, body.Amount)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		api.GET("/payments/methods/:userId", func(c *gin.Context) {
			userID := c.Param("userId")
			cards := svc.Store.SavedCards[userID]
			c.JSON(http.StatusOK, gin.H{"success": true, "methods": cards})
		})

		// ── WALLET & LEDGER ──
		api.GET("/wallet/:userId", func(c *gin.Context) {
			userID := c.Param("userId")
			user, err := svc.GetWallet(userID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"userId":  user.ID,
				"name":    user.Name,
				"role":    user.Role,
				"wallets": gin.H{
					"digital": gin.H{
						"fiatEUR":       user.DigitalWallet.FiatBalanceEUR,
						"loyaltyPoints": user.DigitalWallet.LoyaltyPoints,
						"type":          "DOUBLE_ENTRY_LEDGER_BALANCE",
					},
					"crypto": gin.H{
						"address":    user.CryptoWalletAddress,
						"balanceOEN": "100.00",
						"type":       "NON_CUSTODIAL_ML_DSA_65",
					},
				},
			})
		})

		api.POST("/wallet/deposit", func(c *gin.Context) {
			var body struct {
				UserID string  `json:"userId"`
				Amount float64 `json:"amount"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if body.UserID == "" {
				body.UserID = "user_customer"
			}
			newBal, err := svc.DepositWallet(body.UserID, body.Amount)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success":       true,
				"message":       "Deposit processed successfully",
				"newBalanceEUR": newBal,
			})
		})

		api.POST("/wallet/withdraw", func(c *gin.Context) {
			var body struct {
				UserID string  `json:"userId"`
				Amount float64 `json:"amount"`
				IBAN   string  `json:"iban"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if body.UserID == "" {
				body.UserID = "user_customer"
			}
			newBal, err := svc.WithdrawWallet(body.UserID, body.Amount, body.IBAN)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success":       true,
				"message":       "Withdrawal processed successfully",
				"newBalanceEUR": newBal,
			})
		})

		api.GET("/wallet/ledger", func(c *gin.Context) {
			limitStr := c.DefaultQuery("limit", "50")
			limit, _ := strconv.Atoi(limitStr)
			entries := svc.GetLedgerEntries(limit)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"count":   len(entries),
				"entries": entries,
			})
		})

		// ── COINS ──
		api.GET("/coins/:userId", func(c *gin.Context) {
			userID := c.Param("userId")
			profile := svc.GetCoinProfile(userID)
			c.JSON(http.StatusOK, gin.H{
				"success":      true,
				"userId":       profile.UserID,
				"coinBalance":  profile.CoinBalance,
				"valueEUR":     float64(profile.CoinBalance) / 100.0,
				"totalEarned":  profile.TotalEarned,
				"totalSpent":   profile.TotalSpent,
				"transactions": profile.Transactions,
			})
		})

		api.POST("/coins/redeem", func(c *gin.Context) {
			var body struct {
				UserID   string  `json:"userId"`
				Coins    int64   `json:"coins"`
				Subtotal float64 `json:"subtotal"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if body.UserID == "" {
				body.UserID = "user_customer"
			}
			discount, err := svc.RedeemCoins(body.UserID, body.Coins, body.Subtotal)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success":     true,
				"discountEUR": discount,
				"message":     fmt.Sprintf("Applied €%.2f coin discount", discount),
			})
		})

		// ── CUSTOMER PROFILE & ADDRESSES ──
		api.GET("/customer/profile", func(c *gin.Context) {
			userID := c.DefaultQuery("userId", "user_customer")
			p, err := svc.GetCustomerProfile(userID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "profile": p})
		})

		api.POST("/customer/addresses", func(c *gin.Context) {
			userID := c.DefaultQuery("userId", "user_customer")
			var addr models.SavedAddress
			if err := c.ShouldBindJSON(&addr); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			p, err := svc.AddSavedAddress(userID, addr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"success": true, "profile": p})
		})

		// ── RIDER ──
		api.GET("/rider/profile", func(c *gin.Context) {
			riderID := c.DefaultQuery("riderId", "user_courier")
			p, err := svc.GetRiderProfile(riderID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "rider": p})
		})

		api.POST("/rider/status", func(c *gin.Context) {
			var body struct {
				RiderID  string `json:"riderId"`
				IsOnline bool   `json:"isOnline"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if body.RiderID == "" {
				body.RiderID = "user_courier"
			}
			p, err := svc.ToggleRiderStatus(body.RiderID, body.IsOnline)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "rider": p})
		})

		api.POST("/rider/location", func(c *gin.Context) {
			var body struct {
				RiderID string  `json:"riderId"`
				Lat     float64 `json:"lat"`
				Lng     float64 `json:"lng"`
				Heading float64 `json:"heading"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if body.RiderID == "" {
				body.RiderID = "user_courier"
			}
			loc, err := svc.UpdateRiderLocation(body.RiderID, body.Lat, body.Lng, body.Heading)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "location": loc})
		})

		api.GET("/rider/jobs", func(c *gin.Context) {
			riderID := c.DefaultQuery("riderId", "user_courier")
			jobs := svc.GetRiderJobs(riderID)
			c.JSON(http.StatusOK, gin.H{"success": true, "count": len(jobs), "jobs": jobs})
		})

		// ── ADMIN ──
		api.GET("/admin/dashboard", func(c *gin.Context) {
			stats := svc.GetAdminDashboard()
			c.JSON(http.StatusOK, gin.H{"success": true, "stats": stats})
		})

		api.POST("/admin/commission", func(c *gin.Context) {
			var body struct {
				RatePct float64 `json:"ratePct"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			newRate, err := svc.SetCommissionRate(body.RatePct)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "commissionPct": newRate})
		})

		// ── RECOMMENDATIONS ──
		api.GET("/recommendations", func(c *gin.Context) {
			recs := svc.GetRecommendations()
			c.JSON(http.StatusOK, gin.H{"success": true, "recommendations": recs})
		})

		// ── COUPONS ──
		api.POST("/coupons/validate", func(c *gin.Context) {
			var body struct {
				Code     string  `json:"code"`
				Subtotal float64 `json:"subtotal"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			discount, err := svc.ValidateCoupon(body.Code, body.Subtotal)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "discountEUR": discount})
		})
	}

	return r
}

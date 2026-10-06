package main

import (
	"fmt"
	"log"

	"oengo-api/internal/config"
	"oengo-api/internal/handlers"
	"oengo-api/internal/services"
	"oengo-api/internal/store"
)

func main() {
	cfg := config.LoadConfig()
	st := store.NewStore(cfg)
	svc := services.NewServices(st, cfg)
	router := handlers.SetupRouter(svc)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🍔 OENGO Go Enterprise Backend daemon running on port %s", cfg.Port)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to run OENGO server: %v", err)
	}
}

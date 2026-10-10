package main

import (
	"fmt"
	"log"

	"oengo-api/internal/config"
	"oengo-api/internal/handlers"
	"oengo-api/internal/services"
	"oengo-api/internal/store"
	"oengo-api/internal/db"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	log.Println("Connecting to PostgreSQL database...")
	gormDB, err := gorm.Open(postgres.Open(cfg.DatabaseDSN), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: Failed to connect to DB, continuing with nil DB. Err: %v", err)
	} else {
		log.Println("Database connected successfully. Running migrations...")
		if err := db.AutoMigrateBankingSchema(gormDB); err != nil {
			log.Fatalf("Fatal: Database migration failed: %v", err)
		}
	}

	st := store.NewStore(cfg, gormDB)
	svc := services.NewServices(st, cfg)
	router := handlers.SetupRouter(svc)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🍔 OENGO Go Enterprise Backend daemon running on port %s", cfg.Port)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to run OENGO server: %v", err)
	}
}

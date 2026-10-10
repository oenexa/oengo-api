package db_test

import (
	"context"
	"log"
	"testing"

	"oengo-api/internal/db"
	"oengo-api/internal/models"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestBankingSchemaMigration(t *testing.T) {
	ctx := context.Background()

	// 1. Spin up ephemeral Testcontainer
	pgContainer, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15.3-alpine"),
		tcpostgres.WithDatabase("oengo_test"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %s", err)
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate postgres container: %s", err)
		}
	}()

	// 2. Extract connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %s", err)
	}

	// 3. Connect GORM
	gormDB, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to db: %s", err)
	}

	// 4. Test Professional Banking Schema AutoMigration
	err = db.AutoMigrateBankingSchema(gormDB)
	if err != nil {
		t.Fatalf("failed to run banking schema migration: %s", err)
	}

	// 5. Verify constraints and banking logic
	// Create a user and check uuid generation and defaults
	user := models.BankingUser{
		Name:         "Alice Banking",
		Email:        "alice.bank@example.com",
		PasswordHash: "hashed_pwd",
		Role:         models.RoleCustomer,
	}

	if err := gormDB.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %s", err)
	}

	log.Printf("User successfully created with UUID: %s", user.ID.String())

	// If ID is empty string, the UUID gen didn't work
	if user.ID.String() == "" || user.ID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("Primary key UUID was not generated automatically")
	}
}

package repository

import (
	"context"
	"testing"
	"time"

	"oengo-api/internal/models"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gorm_postgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPostgresUserRepository(t *testing.T) {
	ctx := context.Background()

	// Spin up a PostgreSQL container
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:15.3-alpine",
		tcpostgres.WithDatabase("oengo_test"),
		tcpostgres.WithUsername("oengo_user"),
		tcpostgres.WithPassword("oengo_pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %s", err)
	}

	// Connect GORM to the Testcontainer
	db, err := gorm.Open(gorm_postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to DB: %s", err)
	}

	repo := NewPostgresUserRepository(db)

	// Test 1: Create User
	testUser := &models.User{
		ID:    "user_test_1",
		Name:  "Integration Test User",
		Email: "test@example.com",
		Role:  models.RoleCustomer,
	}

	err = repo.CreateUser(testUser)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Test 2: Fetch User
	fetchedUser, err := repo.GetUserByID("user_test_1")
	if err != nil {
		t.Fatalf("failed to fetch user: %v", err)
	}
	if fetchedUser == nil || fetchedUser.Name != "Integration Test User" {
		t.Fatalf("expected Integration Test User, got %v", fetchedUser)
	}
}

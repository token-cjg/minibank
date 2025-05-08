package db_test

import (
	"os"
	"testing"

	"github.com/token-cjg/minibank/internal/db"
)

func TestNew_NoDatabaseURL(t *testing.T) {
	// Backup and clear DATABASE_URL
	orig := os.Getenv("DATABASE_URL")
	os.Unsetenv("DATABASE_URL")
	defer os.Setenv("DATABASE_URL", orig)

	_, err := db.New()
	if err == nil || err.Error() != "DATABASE_URL not set" {
		t.Fatalf("expected error 'DATABASE_URL not set', got %v", err)
	}
}

func TestNew_InvalidDSN(t *testing.T) {
	orig := os.Getenv("DATABASE_URL")
	defer os.Setenv("DATABASE_URL", orig)

	// Provide an invalid DSN.
	os.Setenv("DATABASE_URL", "invalid-dsn")
	_, err := db.New()
	if err == nil {
		t.Fatal("expected error for invalid DSN")
	}
}

// TestNew_Success is an integration test. It will only run if one sets TEST_DB=true
// and ensure that DATABASE_URL is set to a valid connection string.
func TestNew_Success(t *testing.T) {
	if os.Getenv("TEST_DB") != "true" {
		t.Skip("Skipping integration test because TEST_DB is not true")
	}

	// Assumes DATABASE_URL is set to a valid connection string for testing.
	dbInstance, err := db.New()
	if err != nil {
		t.Fatalf("expected successful connection, got error: %v", err)
	}
	if dbInstance == nil {
		t.Fatal("expected a non-nil db instance")
	}
	dbInstance.Close()
}

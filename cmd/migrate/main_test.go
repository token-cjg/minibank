package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestRunMigration_FileError verifies that runMigration returns an error when the file does not exist.
func TestRunMigration_FileError(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	ctx := context.Background()

	nonExistent := "nonexistent.sql"
	err := runMigration(ctx, db, nonExistent)
	if err == nil || !strings.Contains(err.Error(), "reading file") {
		t.Errorf("expected file reading error, got %v", err)
	}
}

// TestRunMigration_ExecError creates a temporary SQL file and simulates an execution error.
func TestRunMigration_ExecError(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "dummy.sql")
	dummySQL := "SELECT 1;"
	if err := os.WriteFile(filePath, []byte(dummySQL), 0644); err != nil {
		t.Fatalf("failed to write temporary file: %v", err)
	}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	execErr := errors.New("exec error")
	mock.ExpectExec("SELECT 1;").WillReturnError(execErr)

	err = runMigration(ctx, db, filePath)
	if err == nil || !strings.Contains(err.Error(), "executing migration") {
		t.Errorf("expected execution error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestRunMigration_Success creates a temporary SQL file and expects the migration to execute successfully.
func TestRunMigration_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "dummy.sql")
	dummySQL := "SELECT 1;"
	if err := os.WriteFile(filePath, []byte(dummySQL), 0644); err != nil {
		t.Fatalf("failed to write temporary file: %v", err)
	}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	mock.ExpectExec("SELECT 1;").WillReturnResult(sqlmock.NewResult(0, 1))

	err = runMigration(ctx, db, filePath)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestMigrate_Success tests the migrate function by creating a temporary migrations directory with a migration file.
func TestMigrate_Success(t *testing.T) {
	// Create a temporary directory structure.
	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	if err := os.Mkdir(migrationsDir, 0755); err != nil {
		t.Fatalf("failed to create migrations directory: %v", err)
	}

	// Create dummy migration file.
	fileName := "001_dummy.sql"
	filePath := filepath.Join(migrationsDir, fileName)
	dummySQL := "SELECT 1;"
	if err := os.WriteFile(filePath, []byte(dummySQL), 0644); err != nil {
		t.Fatalf("failed to write migration file: %v", err)
	}

	// Change working directory to tmpDir so that migrate() finds the "migrations" folder.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(origDir)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	// Expect the Exec of the dummy SQL.
	mock.ExpectExec("SELECT 1;").WillReturnResult(sqlmock.NewResult(0, 1))

	err = migrate(ctx, db)
	if err != nil {
		t.Errorf("expected no error from migrate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

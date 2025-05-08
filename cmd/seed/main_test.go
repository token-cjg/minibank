package main

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestRunFile_FileNotExist verifies that runFile returns an error when the file does not exist.
func TestRunFile_FileNotExist(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	ctx := context.Background()

	nonExistentPath := filepath.Join(os.TempDir(), "nonexistent.sql")
	err := runFile(ctx, db, nonExistentPath)
	if err == nil || !strings.Contains(err.Error(), "read") {
		t.Errorf("expected file read error, got %v", err)
	}
}

// TestRunFile_EmptyFile creates an empty temporary file and verifies that runFile logs a skip and returns nil.
func TestRunFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyPath := filepath.Join(tmpDir, "empty.sql")
	if err := os.WriteFile(emptyPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write empty file: %v", err)
	}

	// Use a dummy DB – no transaction should occur.
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	// runFile should log and return nil for an empty file.
	if err := runFile(ctx, db, emptyPath); err != nil {
		t.Errorf("expected nil error for empty file, got %v", err)
	}
}

// TestRunFile_Success creates a temporary fixture file with SQL, expecting a transaction commit.
func TestRunFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	seedPath := filepath.Join(tmpDir, "seed.sql")
	// A simple SQL statement.
	dummySQL := "CREATE TABLE test(id INT);"
	if err := os.WriteFile(seedPath, []byte(dummySQL), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	// Expect a Begin, then Exec of dummySQL and Commit.
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(dummySQL)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := runFile(ctx, db, seedPath); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectation: %v", err)
	}
}

// TestRunAll_Success tests that runAll processes only '.sql' files and ignores others.
func TestRunAll_Success(t *testing.T) {
	// Create a temporary directory for fixtures.
	tmpDir := t.TempDir()
	// Create a fake migrations directory.
	fixturesDir := filepath.Join(tmpDir, "fixtures")
	if err := os.Mkdir(fixturesDir, 0755); err != nil {
		t.Fatalf("failed to create fixtures directory: %v", err)
	}

	// Create two SQL files.
	sqlFile1 := filepath.Join(fixturesDir, "001.sql")
	sqlFile2 := filepath.Join(fixturesDir, "002.sql")
	dummySQL1 := "CREATE TABLE a(id INT);"
	dummySQL2 := "CREATE TABLE b(id INT);"
	if err := os.WriteFile(sqlFile1, []byte(dummySQL1), 0644); err != nil {
		t.Fatalf("failed to write sqlFile1: %v", err)
	}
	if err := os.WriteFile(sqlFile2, []byte(dummySQL2), 0644); err != nil {
		t.Fatalf("failed to write sqlFile2: %v", err)
	}
	// Create a non-sql file.
	notSQL := filepath.Join(fixturesDir, "readme.txt")
	if err := os.WriteFile(notSQL, []byte("Ignore me"), 0644); err != nil {
		t.Fatalf("failed to write notSQL: %v", err)
	}

	// Set up sqlmock expectations for both SQL files.
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	// For each sql file in lexical order, we expect a Begin -> Exec -> Commit.
	// Note: files are read in directory order; to be safe, we can check for each SQL.
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(dummySQL1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(dummySQL2)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Change working directory to our temp directory so that runAll finds the "fixtures" folder.
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(origWD)
	}()

	if err := runAll(ctx, db, "fixtures"); err != nil {
		t.Errorf("expected no error from runAll, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations in runAll: %v", err)
	}
}

func TestMainIntegration(t *testing.T) {
	// Prepare a temporary fixtures directory and a seed file.
	tmpDir := t.TempDir()
	fixturesDir := filepath.Join(tmpDir, "fixtures")
	if err := os.Mkdir(fixturesDir, 0755); err != nil {
		t.Fatalf("failed to create fixtures directory: %v", err)
	}
	seedPath := filepath.Join(fixturesDir, "seed.sql")
	dummySQL := "CREATE TABLE integration_test(id INT);"
	if err := os.WriteFile(seedPath, []byte(dummySQL), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}

	// Use sqlmock for DB.
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	// Expect our seed file to be processed.
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(dummySQL)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Temporarily change directory to tmpDir so main() sees the fixtures.
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(origWD)

	// Instead of calling main() (which calls os.Exit on error), call runAll directly.
	if err := runAll(ctx, db, "fixtures"); err != nil {
		t.Errorf("expected no error in integration test, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations not met in integration test: %v", err)
	}
}

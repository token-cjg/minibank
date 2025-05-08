package repo_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/token-cjg/minibank/internal/model"
	"github.com/token-cjg/minibank/internal/repo"
)

func TestCreateAccount(t *testing.T) {
	// Create sqlmock database connection
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock DB: %v", err)
	}
	defer db.Close()

	// Create a new repository instance
	r := repo.New(db)
	ctx := context.Background()
	companyID := int64(1)
	initialBalance := 1000.0

	// Expected inserted account record
	expected := model.Account{
		ID:      10,
		Company: companyID,
		Number:  "1000000000000000",
		Balance: initialBalance,
	}

	// Prepare the expected row result
	rows := sqlmock.NewRows([]string{"account_id", "company_id", "account_number", "account_balance"}).
		AddRow(expected.ID, expected.Company, expected.Number, expected.Balance)

	// Set expectation for the INSERT query
	mock.ExpectQuery(`INSERT INTO account \(company_id, account_balance\) VALUES \(\$1, \$2\)\s+RETURNING account_id, company_id, account_number, account_balance`).
		WithArgs(companyID, initialBalance).
		WillReturnRows(rows)

	// Call CreateAccount
	account, err := r.CreateAccount(ctx, companyID, initialBalance)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check if the returned account matches the expected account
	if account != expected {
		t.Errorf("expected account %+v, got %+v", expected, account)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestListAccountsByCompany(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock DB: %v", err)
	}
	defer db.Close()

	r := repo.New(db)
	ctx := context.Background()
	companyID := int64(1)

	// Create expected rows
	rows := sqlmock.NewRows([]string{"account_id", "company_id", "account_number", "account_balance"}).
		AddRow(1, companyID, 1000000000000000, 500.0).
		AddRow(2, companyID, 1000000000000001, 1500.0)

	// Set expectation for the SELECT query
	mock.ExpectQuery(`SELECT account_id, company_id, account_number, account_balance FROM account WHERE company_id=\$1`).
		WithArgs(companyID).
		WillReturnRows(rows)

	accounts, err := r.ListAccountsByCompany(ctx, companyID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check that we got the correct number of accounts
	if len(accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accounts))
	}

	// Verify the returned data
	expectedFirst := model.Account{ID: 1, Company: companyID, Number: "1000000000000000", Balance: 500.0}
	expectedSecond := model.Account{ID: 2, Company: companyID, Number: "1000000000000001", Balance: 1500.0}

	if accounts[0] != expectedFirst {
		t.Errorf("expected first account %+v, got %+v", expectedFirst, accounts[0])
	}
	if accounts[1] != expectedSecond {
		t.Errorf("expected second account %+v, got %+v", expectedSecond, accounts[1])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetAccountByID(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock DB: %v", err)
	}
	defer db.Close()

	r := repo.New(db)
	ctx := context.Background()
	accountID := int64(1)

	expected := model.Account{
		ID:      accountID,
		Company: 1,
		Number:  "1000000000000000",
		Balance: 750.0,
	}

	// Prepare expected row for GetAccountByID
	rows := sqlmock.NewRows([]string{"account_id", "company_id", "account_number", "account_balance"}).
		AddRow(expected.ID, expected.Company, expected.Number, expected.Balance)

	// Set expectation for the SELECT query
	mock.ExpectQuery(`SELECT account_id, company_id, account_number, account_balance FROM account WHERE account_id=\$1`).
		WithArgs(accountID).
		WillReturnRows(rows)

	account, err := r.GetAccountByID(ctx, accountID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if account != expected {
		t.Errorf("expected account %+v, got %+v", expected, account)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

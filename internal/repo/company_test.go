package repo_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/token-cjg/minibank/internal/model"
	"github.com/token-cjg/minibank/internal/repo"
)

func TestCreateCompany(t *testing.T) {
	// Create sqlmock database connection.
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock DB: %v", err)
	}
	defer db.Close()

	r := repo.New(db)
	ctx := context.Background()
	companyName := "Acme Inc"

	// Expected company result.
	expected := model.Company{
		ID:   1,
		Name: companyName,
	}

	// Prepare expected row result.
	rows := sqlmock.NewRows([]string{"company_id", "company_name"}).
		AddRow(expected.ID, expected.Name)

	// Set expectation for the INSERT query.
	mock.ExpectQuery(`INSERT INTO company \(company_name\) VALUES \(\$1\)\s+RETURNING company_id, company_name`).
		WithArgs(companyName).
		WillReturnRows(rows)

	// Call CreateCompany.
	company, err := r.CreateCompany(ctx, companyName)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Assert results.
	if company != expected {
		t.Errorf("expected company %+v, got %+v", expected, company)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestListCompanies(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock DB: %v", err)
	}
	defer db.Close()

	r := repo.New(db)
	ctx := context.Background()

	// Create expected rows.
	rows := sqlmock.NewRows([]string{"company_id", "company_name"}).
		AddRow(1, "Acme Inc").
		AddRow(2, "Beta Corp")

	// Set expectation for the SELECT query.
	mock.ExpectQuery(`SELECT company_id, company_name FROM company`).
		WillReturnRows(rows)

	companies, err := r.ListCompanies(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify number of companies.
	if len(companies) != 2 {
		t.Fatalf("expected 2 companies, got %d", len(companies))
	}

	expectedFirst := model.Company{ID: 1, Name: "Acme Inc"}
	expectedSecond := model.Company{ID: 2, Name: "Beta Corp"}

	if companies[0] != expectedFirst {
		t.Errorf("expected first company %+v, got %+v", expectedFirst, companies[0])
	}
	if companies[1] != expectedSecond {
		t.Errorf("expected second company %+v, got %+v", expectedSecond, companies[1])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetCompanyByID(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock DB: %v", err)
	}
	defer db.Close()

	r := repo.New(db)
	ctx := context.Background()
	companyID := int64(1)

	expected := model.Company{
		ID:   companyID,
		Name: "Acme Inc",
	}

	// Prepare expected row.
	rows := sqlmock.NewRows([]string{"company_id", "company_name"}).
		AddRow(expected.ID, expected.Name)

	// Set expectation for the SELECT query.
	mock.ExpectQuery(`SELECT company_id, company_name FROM company WHERE company_id=\$1`).
		WithArgs(companyID).
		WillReturnRows(rows)

	company, err := r.GetCompanyByID(ctx, companyID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if company != expected {
		t.Errorf("expected company %+v, got %+v", expected, company)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

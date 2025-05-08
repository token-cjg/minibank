package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/token-cjg/mable-backend-code-test/internal/repo"
)

func newDeps(t *testing.T) (*Account, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	return NewAccount(repo.New(db)), mock
}

func perform(h http.HandlerFunc, method, url string, vars map[string]string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, bytes.NewReader(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if vars != nil {
		req = mux.SetURLVars(req, vars)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestAccountCreate_OK(t *testing.T) {
	h, mock := newDeps(t)

	mock.ExpectQuery(`INSERT INTO account`).
		WithArgs(int64(1), 750.0).
		WillReturnRows(sqlmock.NewRows([]string{
			"account_id", "company_id", "account_number", "account_balance",
		}).AddRow(10, 1, int64(1000000000000010), 750.0))

	body, _ := json.Marshal(map[string]any{"initial_balance": 750.0})
	rec := perform(h.Create, http.MethodPost, "/companies/1/accounts",
		map[string]string{"id": "1"}, body)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rec.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("db expectations: %v", err)
	}
}

func TestAccountCreate_BadJSON(t *testing.T) {
	h, _ := newDeps(t)
	rec := perform(h.Create, http.MethodPost, "/companies/1/accounts",
		map[string]string{"id": "1"}, []byte(`{bad json}`))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestAccountList_Empty(t *testing.T) {
	h, mock := newDeps(t)

	mock.ExpectQuery(`SELECT account_id, company_id, account_number, account_balance FROM account WHERE company_id=\$1`).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{
			"account_id", "company_id", "account_number", "account_balance",
		})) // empty

	rec := perform(h.ListByCompany, http.MethodGet, "/companies/2/accounts",
		map[string]string{"id": "2"}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("db expectations: %v", err)
	}
}

func TestAccountGetByID_OK(t *testing.T) {
	h, mock := newDeps(t)

	mock.ExpectQuery(`SELECT account_id, company_id, account_number, account_balance FROM account WHERE account_id=\$1`).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{
			"account_id", "company_id", "account_number", "account_balance",
		}).AddRow(10, 1, int64(1000000000000010), 500.0))

	rec := perform(h.GetByID, http.MethodGet,
		"/companies/1/accounts/10",
		map[string]string{"id": "10"}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("db expectations: %v", err)
	}
}

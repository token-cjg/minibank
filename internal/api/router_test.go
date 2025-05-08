package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/token-cjg/mable-backend-code-test/internal/repo"
)

func newTestServer(t *testing.T) (*Server, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("cannot create sqlmock: %v", err)
	}
	// repo wraps our fake *sql.DB
	rp := repo.New(db)
	srv := New(rp)
	return srv, mock
}

func perform(t *testing.T, srv *Server, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestCreateCompany(t *testing.T) {
	srv, mock := newTestServer(t)

	// Expect INSERT returning id+name
	mock.ExpectQuery(`INSERT INTO company`).
		WithArgs("Acme Corp").
		WillReturnRows(
			sqlmock.NewRows([]string{"company_id", "company_name"}).
				AddRow(1, "Acme Corp"),
		)

	body, _ := json.Marshal(map[string]string{"company_name": "Acme Corp"})
	rec := perform(t, srv, http.MethodPost, "/companies", body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got status %d, want 201", rec.Code)
	}
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if got["company_name"] != "Acme Corp" {
		t.Fatalf("got body %v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet db expectations: %v", err)
	}
}

func TestListCompanies_Empty(t *testing.T) {
	srv, mock := newTestServer(t)

	mock.ExpectQuery(`SELECT company_id, company_name FROM company`).
		WillReturnRows(sqlmock.NewRows([]string{"company_id", "company_name"})) // no rows

	rec := perform(t, srv, http.MethodGet, "/companies", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d != 200", rec.Code)
	}
	if rec.Body.String() != "[]\n" { // json.Encoder adds newline
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet db expectations: %v", err)
	}
}

func TestCreateAccount(t *testing.T) {
	srv, mock := newTestServer(t)

	// Expect INSERT returning account row
	mock.ExpectQuery(`INSERT INTO account`).
		WithArgs(int64(1), 1000.0).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"account_id", "company_id", "account_number", "account_balance"}).
				AddRow(10, 1, int64(1000000000000001), 1000.0),
		)

	body, _ := json.Marshal(map[string]any{"initial_balance": 1000.0})
	rec := perform(t, srv, http.MethodPost, "/companies/1/accounts", body)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["account_id"] != float64(10) { // json decodes numbers as float64
		t.Fatalf("unexpected resp: %v", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("db expectations: %v", err)
	}
}

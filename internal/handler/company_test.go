package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/token-cjg/mable-backend-code-test/internal/handler"
	"github.com/token-cjg/mable-backend-code-test/internal/repo"
)

func depsCompany(t *testing.T) (*handler.Company, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	return handler.NewCompany(repo.New(db)), mock
}

func call(h http.HandlerFunc, method, url string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, bytes.NewReader(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCompanyCreate_OK(t *testing.T) {
	h, mock := depsCompany(t)

	mock.ExpectQuery(`INSERT INTO company`).
		WithArgs("Acme Corp").
		WillReturnRows(sqlmock.NewRows([]string{"company_id", "company_name"}).
			AddRow(1, "Acme Corp"))

	body, _ := json.Marshal(map[string]string{"company_name": "Acme Corp"})
	rec := call(h.Create, http.MethodPost, "/companies", body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d != 201", rec.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("db expectations: %v", err)
	}
}

func TestCompanyList_Empty(t *testing.T) {
	h, mock := depsCompany(t)

	mock.ExpectQuery(`SELECT company_id, company_name FROM company`).
		WillReturnRows(sqlmock.NewRows([]string{"company_id", "company_name"})) // empty

	rec := call(h.List, http.MethodGet, "/companies", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d != 200", rec.Code)
	}
	if rec.Body.String() != "[]\n" {
		t.Fatalf("want [], got %q", rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("db expectations: %v", err)
	}
}

func TestCompanyGetByID_OK(t *testing.T) {
	h, mock := depsCompany(t)

	mock.ExpectQuery(`SELECT company_id, company_name FROM company WHERE company_id=\$1`).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"company_id", "company_name"}).
			AddRow(2, "Backme Corp"))

	rec := perform(h.GetByID, http.MethodGet, "/companies/2",
		map[string]string{"id": "2"}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d != 200", rec.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("db expectations: %v", err)
	}
}

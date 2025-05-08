package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/token-cjg/mable-backend-code-test/internal/repo"
)

func depsTransfer(t *testing.T) (*Transfer, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	return NewTransfer(repo.New(db)), mock
}

func TestTransferBatch_OK_OneRow(t *testing.T) {
	h, mock := depsTransfer(t)

	// -------------- SQL expectations for a single, valid transfer --------------
	const (
		srcNum int64 = 1000000000000000
		dstNum int64 = 1000000000000001
		srcID  int64 = 1
		dstID  int64 = 2
	)

	mock.ExpectBegin()

	// lock + balance
	mock.ExpectQuery(`SELECT account_id, account_balance.*FOR UPDATE`).
		WithArgs(srcNum).
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "account_balance"}).
			AddRow(srcID, 800.0))

	// target id
	mock.ExpectQuery(`SELECT account_id FROM account WHERE account_number\s*=\s*\$1`).
		WithArgs(dstNum).
		WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow(dstID))

	// debit / credit
	mock.ExpectExec(`UPDATE account SET account_balance = account_balance -`).
		WithArgs(100.0, srcID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE account SET account_balance = account_balance \+`).
		WithArgs(100.0, dstID).WillReturnResult(sqlmock.NewResult(0, 1))

	// insert transaction
	mock.ExpectExec(`INSERT INTO transaction`).
		WithArgs(srcID, dstID, 100.0, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()
	// --------------------------------------------------------------------------

	csvBody := []byte("1000000000000000,1000000000000001,100.00\n")
	req := httptest.NewRequest(http.MethodPost, "/transfer", bytes.NewReader(csvBody))
	req.Header.Set("Content-Type", "text/csv")
	rec := httptest.NewRecorder()
	h.Batch(rec, req)

	println(rec.Body.String())

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d != 204", rec.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("db expectations: %v", err)
	}
}

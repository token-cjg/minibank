package repo_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/token-cjg/minibank/internal/repo"
)

func TestTransfer_Success(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	r := repo.New(db)
	ctx := context.Background()

	// values for a successful transfer where source has sufficient balance
	srcNum := int64(1000000000000000)
	dstNum := int64(1000000000000001)
	amount := 100.0
	srcID := int64(1)
	dstID := int64(2)
	srcBal := 200.0

	// Begin transaction
	mock.ExpectBegin()

	// Query for source account (with lock)
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT account_id, account_balance
           FROM account
          WHERE account_number = $1
          FOR UPDATE`)).
		WithArgs(srcNum).
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "account_balance"}).
			AddRow(srcID, srcBal))
	// Query for target account
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT account_id
           FROM account
          WHERE account_number = $1`)).
		WithArgs(dstNum).
		WillReturnRows(sqlmock.NewRows([]string{"account_id"}).
			AddRow(dstID))
	// Debit source: update account_balance subtracting amount
	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE account
            SET account_balance = account_balance - $1
          WHERE account_id = $2`)).
		WithArgs(amount, srcID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Credit target: update account_balance adding amount
	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE account
            SET account_balance = account_balance + $1
          WHERE account_id = $2`)).
		WithArgs(amount, dstID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Insert transaction record without error message (nil)
	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO transaction
             (source_account_id, target_account_id, transfer_amount, error)
         VALUES ($1,$2,$3,$4)`)).
		WithArgs(srcID, dstID, amount, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// Commit transaction
	mock.ExpectCommit()

	if err := r.Transfer(ctx, srcNum, dstNum, amount); err != nil {
		t.Fatalf("unexpected error during Transfer: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations in Transfer_Success: %v", err)
	}
}

func TestTransfer_Insufficient(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	r := repo.New(db)
	ctx := context.Background()

	// values for transfer with insufficient balance: source balance less than amount
	srcNum := int64(1000000000000000)
	dstNum := int64(1000000000000001)
	amount := 150.0
	srcID := int64(1)
	dstID := int64(2)
	srcBal := 100.0

	mock.ExpectBegin()
	// Query for source account
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT account_id, account_balance
           FROM account
          WHERE account_number = $1
          FOR UPDATE`)).
		WithArgs(srcNum).
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "account_balance"}).
			AddRow(srcID, srcBal))
	// Query for target account
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT account_id
           FROM account
          WHERE account_number = $1`)).
		WithArgs(dstNum).
		WillReturnRows(sqlmock.NewRows([]string{"account_id"}).
			AddRow(dstID))
	// In insufficient scenario, an insert occurs with an error message.
	// Note: Since sqlmock compares pointer equality for non-basic types,
	// construct the expected argument as a pointer.
	msg := "insufficient balance"
	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO transaction
             (source_account_id, target_account_id, transfer_amount, error)
         VALUES ($1,$2,$3,$4)`)).
		WithArgs(srcID, dstID, amount, &msg).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// Commit transaction (even though balance insufficient, Transfer commits)
	mock.ExpectCommit()

	if err := r.Transfer(ctx, srcNum, dstNum, amount); err != nil {
		t.Fatalf("unexpected error during Transfer (insufficient case): %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations in Transfer_Insufficient: %v", err)
	}
}

func TestBatchTransfer_Fatal(t *testing.T) {
	// Test BatchTransfer returns a BatchError if an unexpected error occurs.
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	r := repo.New(db)
	ctx := context.Background()

	// Prepare two transactions: first succeeds, second fails with a fatal error.
	// First transfer (successful).
	srcNum1 := int64(1000000000000000)
	dstNum1 := int64(1000000000000001)
	amount1 := 50.0
	srcID1 := int64(1)
	dstID1 := int64(2)
	srcBal1 := 100.0

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT account_id, account_balance
           FROM account
          WHERE account_number = $1
          FOR UPDATE`)).
		WithArgs(srcNum1).
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "account_balance"}).
			AddRow(srcID1, srcBal1))
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT account_id
           FROM account
          WHERE account_number = $1`)).
		WithArgs(dstNum1).
		WillReturnRows(sqlmock.NewRows([]string{"account_id"}).
			AddRow(dstID1))
	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE account
            SET account_balance = account_balance - $1
          WHERE account_id = $2`)).
		WithArgs(amount1, srcID1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE account
            SET account_balance = account_balance + $1
          WHERE account_id = $2`)).
		WithArgs(amount1, dstID1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO transaction
             (source_account_id, target_account_id, transfer_amount, error)
         VALUES ($1,$2,$3,$4)`)).
		WithArgs(srcID1, dstID1, amount1, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Second transfer: unexpected error during target account query.
	srcNum2 := int64(1000000000000002)
	dstNum2 := int64(1000000000000003)
	amount2 := 75.0
	srcID2 := int64(3)
	srcBal2 := 200.0
	expErr := errors.New("unexpected error")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT account_id, account_balance
           FROM account
          WHERE account_number = $1
          FOR UPDATE`)).
		WithArgs(srcNum2).
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "account_balance"}).
			AddRow(srcID2, srcBal2))
	// For target query, simulate an unexpected error.
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT account_id
           FROM account
          WHERE account_number = $1`)).
		WithArgs(dstNum2).
		WillReturnError(expErr)
	// Expect rollback due to error.
	mock.ExpectRollback()

	// Create batch inputs: first transfer succeeds, second fails.
	txns := []repo.TransferInput{
		{Source: srcNum1, Target: dstNum1, Amount: amount1},
		{Source: srcNum2, Target: dstNum2, Amount: amount2},
	}

	batchErr := r.BatchTransfer(ctx, txns)
	if batchErr == nil {
		t.Fatal("expected BatchTransfer to return error, but got nil")
	}
	// Expect error on second transaction (row index 1).
	if batchErr.Row != 1 {
		t.Errorf("expected error at row 1, got row %d", batchErr.Row)
	}
	if !errors.Is(batchErr.Err, expErr) {
		t.Errorf("expected underlying error %v, got %v", expErr, batchErr.Err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations in BatchTransfer: %v", err)
	}
}

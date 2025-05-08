package repo

import (
	"context"
	"database/sql"
)

type TransferInput struct {
	Source int64
	Target int64
	Amount float64
}

type BatchError struct {
	Row int   // index in csv of the transaction that failed
	Err error // underlying error (e.g. ErrInsufficient, pq error, etc.)
}

func (r *Repo) BatchTransfer(ctx context.Context, txns []TransferInput) *BatchError {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return &BatchError{Row: -1, Err: err}
	}
	defer tx.Rollback()

	for i, t := range txns {
		var (
			srcID, dstID int64
			srcBal       float64
		)
		// 1. lock & fetch source balance + id
		if err := tx.QueryRowContext(ctx,
			`SELECT account_id, account_balance
				FROM account
				WHERE account_number = $1
				FOR UPDATE`,
			t.Source).Scan(&srcID, &srcBal); err != nil {
			return &BatchError{Row: i, Err: err}
		}

		// 2. fetch target id (no lock needed for balance)
		if err := tx.QueryRowContext(ctx,
			`SELECT account_id
				FROM account
				WHERE account_number = $1`,
			t.Target).Scan(&dstID); err != nil {
			return &BatchError{Row: i, Err: err}
		}

		if srcBal < t.Amount {
			return &BatchError{Row: i, Err: ErrInsufficient}
		}

		// 3. debit / credit using the **ids**
		if _, err := tx.ExecContext(ctx,
			`UPDATE account
				SET account_balance = account_balance - $1
				WHERE account_id = $2`,
			t.Amount, srcID); err != nil {
			return &BatchError{Row: i, Err: err}
		}

		if _, err := tx.ExecContext(ctx,
			`UPDATE account
				SET account_balance = account_balance + $1
				WHERE account_id = $2`,
			t.Amount, dstID); err != nil {
			return &BatchError{Row: i, Err: err}
		}

		// 4. record the transfer
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO transaction
				(source_account_id, target_account_id, transfer_amount)
				VALUES ($1, $2, $3)`,
			srcID, dstID, t.Amount); err != nil {
			return &BatchError{Row: i, Err: err}
		}

	}
	return wrapErr(tx.Commit())
}

func wrapErr(err error) *BatchError {
	if err == nil {
		return nil
	}
	return &BatchError{Row: -1, Err: err}
}

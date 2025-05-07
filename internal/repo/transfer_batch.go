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

	debit := `UPDATE account SET account_balance = account_balance - $1 WHERE account_id=$2`
	credit := `UPDATE account SET account_balance = account_balance + $1 WHERE account_id=$2`
	insert := `INSERT INTO transaction (source_account_id,target_account_id,transfer_amount) VALUES ($1,$2,$3)`

	for i, t := range txns {
		var srcBal float64
		if err := tx.QueryRowContext(ctx,
			`SELECT account_balance FROM account WHERE account_id=$1 FOR UPDATE`,
			t.Source).Scan(&srcBal); err != nil {
			return &BatchError{Row: i, Err: err}
		}
		if srcBal < t.Amount {
			return &BatchError{Row: i, Err: ErrInsufficient}
		}
		if _, err := tx.ExecContext(ctx, debit, t.Amount, t.Source); err != nil {
			return &BatchError{Row: i, Err: err}
		}
		if _, err := tx.ExecContext(ctx, credit, t.Amount, t.Target); err != nil {
			return &BatchError{Row: i, Err: err}
		}
		if _, err := tx.ExecContext(ctx, insert, t.Source, t.Target, t.Amount); err != nil {
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

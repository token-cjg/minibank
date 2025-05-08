package repo

import (
	"context"
	"database/sql"
	"errors"
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
	for i, t := range txns {
		if err := r.Transfer(ctx, t.Source, t.Target, t.Amount); err != nil {
			// only treat *unexpected* DB errors as fatal
			if !errors.Is(err, ErrInsufficient) {
				return &BatchError{Row: i, Err: err}
			}
		}
	}
	return nil
}

func (r *Repo) Transfer(ctx context.Context, srcNum, dstNum int64, amount float64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		srcID, dstID int64
		srcBal       float64
	)

	// lock + fetch source id & balance
	if err := tx.QueryRowContext(ctx,
		`SELECT account_id, account_balance
		   FROM account
		  WHERE account_number = $1
		  FOR UPDATE`,
		srcNum).Scan(&srcID, &srcBal); err != nil {
		return err
	}

	// fetch target id
	if err := tx.QueryRowContext(ctx,
		`SELECT account_id
		   FROM account
		  WHERE account_number = $1`,
		dstNum).Scan(&dstID); err != nil {
		return err
	}

	if srcBal < amount {
		msg := "insufficient balance"
		_ = r.insertTx(ctx, tx, srcID, dstID, amount, &msg)
		return tx.Commit()
	}

	// debit / credit using account_id
	if _, err := tx.ExecContext(ctx,
		`UPDATE account
		    SET account_balance = account_balance - $1
		  WHERE account_id = $2`,
		amount, srcID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE account
		    SET account_balance = account_balance + $1
		  WHERE account_id = $2`,
		amount, dstID); err != nil {
		return err
	}

	if err := r.insertTx(ctx, tx, srcID, dstID, amount, nil); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repo) insertTx(ctx context.Context, q execer,
	srcID, dstID int64, amount float64, errMsg *string) error {
	_, err := q.ExecContext(ctx,
		`INSERT INTO transaction
             (source_account_id, target_account_id, transfer_amount, error)
         VALUES ($1,$2,$3,$4)`,
		srcID, dstID, amount, errMsg)
	return err
}

type execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

package repo

import (
	"context"
	"database/sql"
)

func (r *Repo) Transfer(ctx context.Context, srcID, dstID int64, amount float64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var srcBal float64
	if err := tx.QueryRowContext(ctx,
		`SELECT account_balance FROM account WHERE account_number=$1 FOR UPDATE`, srcID).
		Scan(&srcBal); err != nil {
		return err
	}

	if srcBal < amount {
		return ErrInsufficient
	}

	// Debit & credit
	if _, err := tx.ExecContext(ctx,
		`UPDATE account SET account_balance = account_balance - $1 WHERE account_number=$2`, amount, srcID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE account SET account_balance = account_balance + $1 WHERE account_number=$2`, amount, dstID); err != nil {
		return err
	}

	// Record transaction
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO transaction (source_account_id,target_account_id,transfer_amount) VALUES ($1,$2,$3)`,
		srcID, dstID, amount); err != nil {
		return err
	}

	return tx.Commit()
}

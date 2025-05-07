package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/token-cjg/mable-backend-code-test/internal/model"
)

type Repo struct{ db *sql.DB }

func New(db *sql.DB) *Repo { return &Repo{db} }

/* --------- Company --------- */

func (r *Repo) CreateCompany(ctx context.Context, name string) (model.Company, error) {
	var c model.Company
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO company (company_name) VALUES ($1) RETURNING company_id, company_name`,
		name).Scan(&c.ID, &c.Name)
	return c, err
}

func (r *Repo) ListCompanies(ctx context.Context) ([]model.Company, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT company_id, company_name FROM company`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Company
	for rows.Next() {
		var c model.Company
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *Repo) GetCompanyByID(ctx context.Context, companyID int64) (model.Company, error) {
	var c model.Company
	err := r.db.QueryRowContext(ctx,
		`SELECT company_id, company_name FROM company WHERE company_id=$1`,
		companyID).Scan(&c.ID, &c.Name)
	return c, err
}

/* --------- Account --------- */

func (r *Repo) CreateAccount(ctx context.Context, companyID int64, balance float64) (model.Account, error) {
	var a model.Account
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO account (company_id, account_balance) VALUES ($1, $2)
		 RETURNING account_id, company_id, account_balance`,
		companyID, balance).
		Scan(&a.ID, &a.Company, &a.Balance)
	return a, err
}

func (r *Repo) ListAccountsByCompany(ctx context.Context, companyID int64) ([]model.Account, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT account_id, company_id, account_balance FROM account WHERE company_id=$1`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accs := []model.Account{}
	for rows.Next() {
		var a model.Account
		if err := rows.Scan(&a.ID, &a.Company, &a.Balance); err != nil {
			return nil, err
		}
		accs = append(accs, a)
	}
	return accs, rows.Err()
}

func (r *Repo) GetAccountByID(ctx context.Context, accountID int64) (model.Account, error) {
	var a model.Account
	err := r.db.QueryRowContext(ctx,
		`SELECT account_id, company_id, account_balance FROM account WHERE account_id=$1`,
		accountID).Scan(&a.ID, &a.Company, &a.Balance)
	return a, err
}

/* --------- Transfer (business txn) --------- */

// ErrInsufficient funds sentinel
var ErrInsufficient = errors.New("insufficient balance")

func (r *Repo) Transfer(ctx context.Context, srcID, dstID int64, amount float64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var srcBal float64
	if err := tx.QueryRowContext(ctx,
		`SELECT account_balance FROM account WHERE account_id=$1 FOR UPDATE`, srcID).
		Scan(&srcBal); err != nil {
		return err
	}

	if srcBal < amount {
		return ErrInsufficient
	}

	// Debit & credit
	if _, err := tx.ExecContext(ctx,
		`UPDATE account SET account_balance = account_balance - $1 WHERE account_id=$2`, amount, srcID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE account SET account_balance = account_balance + $1 WHERE account_id=$2`, amount, dstID); err != nil {
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

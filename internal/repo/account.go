package repo

import (
	"context"

	"github.com/token-cjg/mable-backend-code-test/internal/model"
)

func (r *Repo) CreateAccount(ctx context.Context, companyID int64, balance float64) (model.Account, error) {
	var a model.Account
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO account (company_id, account_balance) VALUES ($1, $2)
                RETURNING account_id, company_id, account_number, account_balance`,
		companyID, balance).Scan(&a.ID, &a.Company, &a.Number, &a.Balance)
	return a, err
}

func (r *Repo) ListAccountsByCompany(ctx context.Context, companyID int64) ([]model.Account, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT account_id, company_id, account_number, account_balance FROM account WHERE company_id=$1`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accs := []model.Account{}
	for rows.Next() {
		var a model.Account
		if err := rows.Scan(&a.ID, &a.Company, &a.Number, &a.Balance); err != nil {
			return nil, err
		}
		accs = append(accs, a)
	}
	return accs, rows.Err()
}

func (r *Repo) GetAccountByID(ctx context.Context, accountID int64) (model.Account, error) {
	var a model.Account
	err := r.db.QueryRowContext(ctx,
		`SELECT account_id, company_id, account_number, account_balance FROM account WHERE account_id=$1`,
		accountID).Scan(&a.ID, &a.Company, &a.Number, &a.Balance)
	return a, err
}

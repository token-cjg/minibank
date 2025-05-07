package repo

import (
	"context"

	"github.com/token-cjg/mable-backend-code-test/internal/model"
)

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

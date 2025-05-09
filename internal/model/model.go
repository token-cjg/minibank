// Package model contains the data structures used in the application.
// It is used to define the data models for the database entities.
// The models are used in the repository layer to interact with the database.
package model

type Company struct {
	ID   int64  `json:"company_id"`
	Name string `json:"company_name"`
}

type Account struct {
	ID      int64   `json:"account_id"`
	Company int64   `json:"company_id"`
	Number  string  `json:"account_number"`
	Balance float64 `json:"account_balance"`
}

type Transaction struct {
	ID        int64   `json:"tx_id"`
	Source    int64   `json:"source_account_id"`
	Target    int64   `json:"target_account_id"`
	Amount    float64 `json:"transfer_amount"`
	Error     *string `json:"error,omitempty"`
	CreatedAt string  `json:"created_at"`
}

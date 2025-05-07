package model

type Company struct {
	ID   int64  `json:"company_id"`
	Name string `json:"company_name"`
}

type Account struct {
	ID      int64   `json:"account_id"`
	Company int64   `json:"company_id"`
	Balance float64 `json:"account_balance"`
}

type Transaction struct {
	ID        int64   `json:"tx_id"`
	Source    int64   `json:"source_account_id"`
	Target    int64   `json:"target_account_id"`
	Amount    float64 `json:"transfer_amount"`
	CreatedAt string  `json:"created_at"`
}

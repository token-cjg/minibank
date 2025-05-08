// Package handler provides HTTP handlers for managing accounts and transactions.
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/token-cjg/mable-backend-code-test/internal/model"
	"github.com/token-cjg/mable-backend-code-test/internal/repo"
)

type Account struct{ Repo *repo.Repo }

func NewAccount(r *repo.Repo) *Account { return &Account{Repo: r} }

/*
Create is a handler for creating a new account.

	POST /companies/{id}/accounts
	Content-Type: application/json
	Body: {"initial_balance": 1000.0}
*/
func (h *Account) Create(w http.ResponseWriter, r *http.Request) {
	companyID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Error(w, "bad company id", http.StatusBadRequest)
		return
	}
	var req struct {
		Balance float64 `json:"initial_balance"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	acct, err := h.Repo.CreateAccount(r.Context(), companyID, req.Balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(acct)
}

/*
	 ListByCompany is a handler for listing all accounts for a company.
		GET /companies/{id}/accounts
*/
func (h *Account) ListByCompany(w http.ResponseWriter, r *http.Request) {
	companyID, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	accs, err := h.Repo.ListAccountsByCompany(r.Context(), companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if accs == nil {
		accs = []model.Account{}
	}
	json.NewEncoder(w).Encode(accs)
}

/*
GetByID is a handler for getting an account by its ID.

	GET /companies/{id}/accounts/{id}
*/
func (h *Account) GetByID(w http.ResponseWriter, r *http.Request) {
	accountID, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	acc, err := h.Repo.GetAccountByID(r.Context(), accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(acc)
}

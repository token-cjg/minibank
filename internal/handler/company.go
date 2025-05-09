package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/token-cjg/minibank/internal/model"
	"github.com/token-cjg/minibank/internal/repo"
)

type Company struct{ Repo *repo.Repo }

func NewCompany(r *repo.Repo) *Company { return &Company{Repo: r} }

/*
	 Create is a handler for creating a new company.
			POST /companies
			Content-Type: application/json
			Body: {"company_name": "Company Name"}
*/
func (h *Company) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"company_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := h.Repo.CreateCompany(r.Context(), req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

/*
	 List is a handler for listing all companies.
			GET /companies
*/
func (h *Company) List(w http.ResponseWriter, r *http.Request) {
	cs, err := h.Repo.ListCompanies(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if cs == nil {
		cs = []model.Company{}
	}
	writeJSON(w, http.StatusOK, cs)
}

/*
	 GetByID is a handler for getting a company by ID.
			GET /companies/{id}
*/
func (h *Company) GetByID(w http.ResponseWriter, r *http.Request) {
	companyID, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	c, err := h.Repo.GetCompanyByID(r.Context(), companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

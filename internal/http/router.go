package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/token-cjg/mable-backend-code-test/internal/repo"
)

type Server struct {
	r   *mux.Router
	rep *repo.Repo
}

func NewServer(rep *repo.Repo) *Server {
	s := &Server{r: mux.NewRouter().StrictSlash(true), rep: rep}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.r.HandleFunc("/companies", s.createCompany).Methods(http.MethodPost)
	s.r.HandleFunc("/companies", s.listCompanies).Methods(http.MethodGet)
	s.r.HandleFunc("/companies/{id:[0-9]+}/accounts", s.createAccount).Methods(http.MethodPost)
	s.r.HandleFunc("/companies/{id:[0-9]+}/accounts", s.listAccounts).Methods(http.MethodGet)
	s.r.HandleFunc("/transfer", s.transfer).Methods(http.MethodPost)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.r.ServeHTTP(w, r) }

/* -------- Handlers ---------- */

func (s *Server) createCompany(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"company_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	c, err := s.rep.CreateCompany(r.Context(), req.Name)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(c)
}

func (s *Server) listCompanies(w http.ResponseWriter, r *http.Request) {
	cs, err := s.rep.ListCompanies(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(cs)
}

func (s *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	companyID, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	var req struct {
		Balance float64 `json:"initial_balance"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	a, err := s.rep.CreateAccount(r.Context(), companyID, req.Balance)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(a)
}

func (s *Server) listAccounts(w http.ResponseWriter, r *http.Request) {
	companyID, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	accs, err := s.rep.ListAccountsByCompany(r.Context(), companyID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(accs)
}

func (s *Server) transfer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Source int64   `json:"source_account_id"`
		Target int64   `json:"target_account_id"`
		Amount float64 `json:"transfer_amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := s.rep.Transfer(r.Context(), req.Source, req.Target, req.Amount); err != nil {
		code := 500
		if err == repo.ErrInsufficient {
			code = 409
		}
		http.Error(w, err.Error(), code)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

package http

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
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
	// 1. Quick guard
	if ct := r.Header.Get("Content-Type"); ct != "text/csv" && ct != "text/plain" {
		http.Error(w, "expecting Content-Type: text/csv", 415)
		return
	}

	// 2. Stream‑parse rows
	csvr := csv.NewReader(r.Body)
	csvr.FieldsPerRecord = 3
	defer r.Body.Close()

	var batch []repo.TransferInput
	line := 0
	for {
		rec, err := csvr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("bad CSV on line %d: %v", line+1, err), 400)
			return
		}

		src, err1 := strconv.ParseInt(rec[0], 10, 64)
		dst, err2 := strconv.ParseInt(rec[1], 10, 64)
		amt, err3 := strconv.ParseFloat(rec[2], 64)
		if err := firstErr(err1, err2, err3); err != nil {
			http.Error(w, fmt.Sprintf("parse error on line %d: %v", line+1, err), 400)
			return
		}
		batch = append(batch, repo.TransferInput{Source: src, Target: dst, Amount: amt})
		line++
	}

	// 3. Run batch
	if berr := s.rep.BatchTransfer(r.Context(), batch); berr != nil {
		status := 500
		if berr.Err == repo.ErrInsufficient {
			status = 409
		}
		resp := map[string]any{
			"error": berr.Err.Error(),
			"row":   berr.Row + 1,
		}
		writeJSON(w, status, resp)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v) // ignore encode error -> connection already closed
}

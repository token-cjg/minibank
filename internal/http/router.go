package http

import (
	"net/http"

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

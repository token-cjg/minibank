// Package api provides the HTTP server and routing for the application.
// It uses the Gorilla Mux router for handling HTTP requests and responses.
// The server is responsible for defining the API endpoints and their handlers.
// It also sets up the necessary middleware for logging and error handling.
package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/token-cjg/minibank/internal/handler"
	"github.com/token-cjg/minibank/internal/repo"
)

type Server struct {
	router *mux.Router
}

func New(rep *repo.Repo) *Server {
	s := &Server{router: mux.NewRouter().StrictSlash(true)}

	account := handler.NewAccount(rep)
	company := handler.NewCompany(rep)
	transfer := handler.NewTransfer(rep)

	s.router.HandleFunc("/companies", company.Create).Methods(http.MethodPost)
	s.router.HandleFunc("/companies", company.List).Methods(http.MethodGet)
	s.router.HandleFunc("/companies/{id:[0-9]+}", company.GetByID).Methods(http.MethodGet)

	s.router.HandleFunc("/companies/{id:[0-9]+}/accounts",
		account.Create).Methods(http.MethodPost)
	s.router.HandleFunc("/companies/{id:[0-9]+}/accounts",
		account.ListByCompany).Methods(http.MethodGet)
	s.router.HandleFunc("/companies/{id:[0-9]+}/accounts/{id:[0-9]+}",
		account.GetByID).Methods(http.MethodGet)

	s.router.HandleFunc("/transfer", transfer.Batch).Methods(http.MethodPost)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

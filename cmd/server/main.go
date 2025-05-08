package main

import (
	"log"
	"net/http"
	"time"

	"github.com/token-cjg/minibank/internal/api"
	"github.com/token-cjg/minibank/internal/db"
	"github.com/token-cjg/minibank/internal/repo"
)

func main() {
	pg, err := db.New()
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pg.Close()

	rep := repo.New(pg)
	srv := api.New(rep)

	server := newHTTPServer(srv)
	log.Println("🚀  listening on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func newHTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

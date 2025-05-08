package main

import (
	"log"
	"net/http"
	"time"

	"github.com/token-cjg/mable-backend-code-test/internal/api"
	"github.com/token-cjg/mable-backend-code-test/internal/db"
	"github.com/token-cjg/mable-backend-code-test/internal/repo"
)

func main() {
	pg, err := db.New()
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pg.Close()

	rep := repo.New(pg)
	srv := api.New(rep)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      srv,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("🚀  listening on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

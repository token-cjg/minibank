package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// healthHandler is a simple “ping” endpoint.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	// Router setup
	r := mux.NewRouter().StrictSlash(true)

	// Routes – add more here
	r.HandleFunc("/health", healthHandler).Methods(http.MethodGet)

	// Optional: global middleware example
	r.Use(loggingMiddleware)

	// Server
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("🚀  listening on http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

// loggingMiddleware prints each request in Apache‑style combined format.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			log.Printf("%s %s %s %dms",
				r.Method,
				r.RequestURI,
				r.UserAgent(),
				time.Since(start).Milliseconds(),
			)
		}()
		next.ServeHTTP(w, r)
	})
}

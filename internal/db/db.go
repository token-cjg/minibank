// Package db provides a database connection for the application.
// It uses the pgx driver for PostgreSQL and sets up connection pooling.
// The connection pool settings are configured to optimize performance.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func New() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	// Set connection pool settings
	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(10 * time.Minute)

	return db, db.Ping()
}

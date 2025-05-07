package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx := context.Background()
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := migrate(ctx, db); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}
}

func migrate(ctx context.Context, db *sql.DB) error {
	migrationsDir := "migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("reading migrations directory: %v", err)
	}

	for _, file := range files {
		if err := runMigration(ctx, db, filepath.Join(migrationsDir, file.Name())); err != nil {
			return fmt.Errorf("running migration %s: %v", file.Name(), err)
		}
	}
	return nil
}

func runMigration(ctx context.Context, db *sql.DB, file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading file: %v", err)
	}

	if _, err := db.ExecContext(ctx, string(data)); err != nil {
		return fmt.Errorf("executing migration: %v", err)
	}
	return nil
}

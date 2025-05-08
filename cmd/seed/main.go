// cmd/seed/main.go
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// CLI flags
	dir := flag.String("dir", "fixtures", "directory containing seed SQL files")
	file := flag.String("file", "seed.sql", "single seed file (ignored if --all is set)")
	all := flag.Bool("all", false, "run every *.sql file in --dir in lexical order")
	flag.Parse()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL env var not set")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if *all {
		if err := runAll(ctx, db, *dir); err != nil {
			log.Fatalf("seed: %v", err)
		}
	} else {
		seedPath := filepath.Join(*dir, *file)
		if err := runFile(ctx, db, seedPath); err != nil {
			log.Fatalf("seed: %v", err)
		}
	}
	log.Println("✅  seed completed")
}

func runAll(ctx context.Context, db *sql.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".sql" {
			continue
		}
		p := filepath.Join(dir, e.Name())
		if err := runFile(ctx, db, p); err != nil {
			return err
		}
	}
	return nil
}

func runFile(ctx context.Context, db *sql.DB, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if len(data) == 0 {
		log.Printf("skip empty file %s", path)
		return nil
	}

	// Wrap in a transaction so fixture load is atomic
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, string(data)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("exec %s: %w", path, err)
	}
	return tx.Commit()
}

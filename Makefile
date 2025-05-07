.PHONY: db_create db_drop run db_migrate all

db_create:
	@echo "Creating bank database..."
	psql -U postgres -d postgres -c "CREATE DATABASE bank;" || echo "Database already exists."

db_drop:
	@echo "Dropping bank database..."
	psql -U postgres -d postgres -c "DROP DATABASE IF EXISTS bank;" || echo "Database does not exist."

run:
	DATABASE_URL="postgres://postgres:postgres@localhost:5432/bank?sslmode=disable" go run ./cmd/server

db_migrate:
	DATABASE_URL="postgres://postgres:postgres@localhost:5432/bank?sslmode=disable" go run ./cmd/migrate

all: db_create db_migrate run

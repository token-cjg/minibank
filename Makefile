.PHONY: db_create db_drop run db_migrate db_seed test coverage docs serve_docs all

db_create:
	@echo "Creating bank database..."
	psql -U postgres -d postgres -c "CREATE DATABASE bank;" || echo "Database already exists."

db_drop:
	@echo "Dropping bank database..."
	psql -U postgres -d postgres -c "DROP DATABASE IF EXISTS bank;" || echo "Database does not exist."


db_migrate:
	DATABASE_URL="postgres://postgres:postgres@localhost:5432/bank?sslmode=disable" go run ./cmd/migrate

db_seed:
	DATABASE_URL="postgres://postgres:postgres@localhost:5432/bank?sslmode=disable" go run ./cmd/seed -dir ./fixtures -file seed.sql

run:
	DATABASE_URL="postgres://postgres:postgres@localhost:5432/bank?sslmode=disable" go run ./cmd/server

test:
	go test -v ./... -coverprofile=coverage.out
	@echo "Running tests..."

coverage:
	go test -v ./... -coverprofile=coverage.out
	@total=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	threshold=60; \
	echo "Total Coverage: $$total% (min required is $$threshold%)"; \
	if [ $$(echo "$$total < $$threshold" | bc -l) -eq 1 ]; then \
		echo "Coverage below threshold!"; exit 1; \
	fi

lint:
	@echo "Checking if golangci-lint is installed, if not, installing..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		echo "Adding golangci-lint to PATH..."; \
		echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bash_profile; \
		source ~/.bash_profile; \
		echo "golangci-lint installed successfully!"; \
    }
	@echo "Running linters..."
	golangci-lint run ./... --config .golangci.yml

docs:
	@echo "Checking if go-docs is installed, if not, installing..."
	@command -v gomarkdoc >/dev/null 2>&1 || { \
		echo "Installing gomarkdoc..."; \
		go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest; \
		echo "gomarkdoc installed successfully!"; \
	}
	@echo "📝  generating Markdown docs ➜ docs/api.md"
	@mkdir -p docs
	@gomarkdoc --output docs/api.md ./...

serve_docs:
	@echo "🌐 browse http://localhost:6060/github.com/token-cjg/minibank"
	@command -v pkgsite >/dev/null 2>&1 || { \
		echo "Installing pkgsite..."; \
		go install golang.org/x/pkgsite/cmd/pkgsite@latest; \
		echo "pkgsite installed successfully!"; \
	}
	@echo "Running pkgsite..."
	@echo "You can view the documentation at http://localhost:6060/github.com/token-cjg/minibank"
	@pkgsite -http "localhost:6060" -open

lint_fix:
	@echo "Running linters with auto-fix..."
	golangci-lint run --fix ./... --config .golangci.yml
	@echo "Linters completed successfully!"

all: db_create db_migrate db_seed test coverage lint
	@echo "All tasks completed successfully!"
	@echo "You can now run the server with 'make run'."

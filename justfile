# justfile for HR Backend

# List all available commands by default
default:
	@just --list

# Core setup: Install tools and initialize Lefthook
setup:
	@echo "Installing tools and setting up git hooks..."
	go install github.com/evilmartians/lefthook@latest
	go install github.com/air-verse/air@latest
	lefthook install

# Run static analysis (linter)
lint:
	golangci-lint run ./...

# Run tests
test:
	go test -v ./...

# Generate SQLC code
generate:
	sqlc generate


# Run development server with hot module replacement using Air
dev:
	air -c .air.toml

# Build the project
build:
	go build -o bin/server main.go

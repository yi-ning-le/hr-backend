# HR Backend API

A Go backend service for HR management built with **Gin + sqlc + pgx**.

## Tech Stack

- **Gin** - HTTP web framework
- **sqlc** - Type-safe SQL query builder
- **pgx/v5** - PostgreSQL driver and connection pool
- **PostgreSQL** - Database

## Project Structure

```
.
├── cmd/api/              # Application entrypoints
├── internal/             # Private application code
│   ├── config/          # Configuration management
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── model/           # Data models and structs
│   ├── repository/      # Data access layer (sqlc generated)
│   │   └── query/       # SQL queries for sqlc
│   ├── service/         # Business logic
│   └── utils/           # Utility functions
├── pkg/                 # Public library code
│   ├── database/        # Database utilities (pgx)
│   └── logger/          # Logging utilities
├── migrations/          # Database schema files
├── sqlc.yaml           # sqlc configuration
├── docs/               # Documentation
├── scripts/            # Build and deployment scripts
└── test/               # Test files
```

## Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL
- sqlc CLI (for development)

### Installation

1. Clone the repository
2. Set up the development environment (installs tools and git hooks):

   ```bash
   make setup
   ```

3. Set up environment variables:

   ```bash
   export SERVER_PORT=8080
   export DATABASE_URL=postgres://localhost/hrdb?sslmode=disable
   ```

4. Generate code from SQL queries:

   ```bash
   sqlc generate
   ```

5. Run the application:

   ```bash
   make run
   ```

   **OR for development with hot module replacement (HMR):**

   ```bash
   make dev
   ```

### API Endpoints

- `GET /health` - Health check
- `GET /api/v1/users` - Get users

## Development

### Database Changes

1. Update SQL schema in `migrations/`
2. Update SQL queries in `internal/repository/query/`
3. Generate Go code:
   ```bash
   sqlc generate
   ```

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o bin/hr-backend main.go
```

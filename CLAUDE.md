# Claude Configuration: HR Backend

Project: HR Management System - Backend API
Stack: Go 1.26+, Gin, PostgreSQL (pgx), sqlc
Description: Backend service for an HR management system with clean architecture.

## Core Commands
- **Build**: `just build` (Outputs to `bin/server`)
- **Dev**: `just dev` (with hot module replacement)
- **Test**: `just test`
- **Migrate**: `just migrate`
- **Generate SQL**: `just generate` (Run after modifying SQL in `migrations/` or `internal/repository/query/`)
- **Lint**: `just lint`

## Project Structure
```
main.go                 # API server entry
cmd/migrate/            # Migration tool
internal/               # Application code
├── handler/            # HTTP handlers
├── service/            # Business logic
├── repository/         # Data access (sqlc-generated)
├── config/             # Configuration
migrations/             # SQL schemas
```

## Key Rules
- **TDD Mandatory**: Tests must fail before implementation
- **Use sqlc** for all database access (no raw SQL in handlers/services)
- **Inject dependencies** via interfaces for testability

## Detailed Configuration
See [AGENT.md](./AGENT.md) for complete project configuration, including:
- Detailed skill matrix
- Full architecture layers
- Development conventions
- Database migration workflows

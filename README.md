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
│   ├── repository/      # Data access layer (just generated)
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

- Go 1.26+
- PostgreSQL
- sqlc CLI (for development)

### Installation

1. Clone the repository
2. Set up the development environment (installs tools and git hooks):

   ```bash
   just setup
   ```

3. Set up environment variables:

   ```bash
   export SERVER_PORT=8080
   export DATABASE_URL=postgres://localhost/hrdb?sslmode=disable
   ```

4. Generate code from SQL queries:

   ```bash
   just generate
   ```

5. Run the development server (with hot module replacement):

   ```bash
   just dev
   ```

### API Endpoints

- `GET /health` - Health check
- `POST /candidates` - Create candidate with required resume upload (`multipart/form-data`)
  - Form field `file`: PDF only, max 10MB
  - Form field `data`: JSON string for candidate fields (`name`, `email`, `phone`, `experienceYears`, `education`, `appliedJobId`, `channel`, `appliedAt`, etc.)

## Development

### Database Changes

1. Update SQL schema in `migrations/`
2. Run migrations:
   ```bash
   just migrate
   ```
3. Update SQL queries in `internal/repository/query/`
4. Generate Go code:
   ```bash
   just generate
   ```

To re-run a specific migration (drops tables, removes record, re-applies):

```bash
just remigrate <migration_file.sql>
```

Both `migrate` and `remigrate` commands trigger hot reload if Air is running.

### Running Tests

```bash
just test
```

### Building

```bash
just build
```

# ParaSempre - Backend

Wedding guest management API built in Go.

## Tech Stack

- **Go 1.26** with stdlib `net/http` (enhanced ServeMux with method routing)
- **pgx/v5** for PostgreSQL (Supabase) — no ORM
- **excelize/v2** for XLSX import, stdlib `encoding/csv` for CSV
- **log/slog** for structured logging

## Architecture

Three-layer domain-oriented structure: **Handler → Service → Repository**.

```
backend/
  cmd/server/main.go          # Entrypoint, CORS middleware, graceful shutdown
  internal/
    config/config.go           # Env-based config (DATABASE_URL, PORT, CORS_ORIGIN)
    database/database.go       # pgxpool connection helper
    guest/                     # Single domain package (not split by layer)
      model.go                 # Guest, CreateGuestInput, UpdateGuestInput
      repository.go            # Repository interface
      repository_postgres.go   # pgx implementation
      service.go               # Business logic + validation
      handler.go               # HTTP handlers + route registration
      importer.go              # CSV/XLSX parsing (io.Reader based)
      *_test.go                # Tests for service, handler, importer
  migrations/
    001_create_guests.sql
```

## Key Patterns

- **Repository is an interface** — tests use a mock with function fields (no codegen)
- **Pointer fields** in `UpdateGuestInput` (`*string`, `*bool`) for partial updates via `COALESCE`
- **Parsers accept `io.Reader`** — no file path coupling, fully testable
- **Table-driven tests** with `testing` stdlib only (no testify)
- **CORS middleware** is a single function in `main.go`
- **UUID primary keys** via Postgres `gen_random_uuid()`

## Commands

```bash
make test    # Run all unit tests
make run     # Start server (loads .env automatically)
make build   # Build binary to bin/server
make migrate # Run SQL migrations against Supabase
```

## API Endpoints

| Method | Path                 | Description          |
|--------|----------------------|----------------------|
| GET    | /api/guests          | List all guests      |
| POST   | /api/guests          | Create guest         |
| GET    | /api/guests/{id}     | Get guest by ID      |
| PUT    | /api/guests/{id}     | Update guest         |
| DELETE | /api/guests/{id}     | Delete guest         |
| POST   | /api/guests/import   | Import CSV/XLSX      |

## Supabase Connection

We connect **directly to PostgreSQL via pgx** — no Supabase REST API, no API keys needed.

- Use the **direct connection** string from Supabase Dashboard (Settings > Database > Connection string > URI)
- Must use port **5432** (direct), NOT 6543 (transaction pooler) — pgx uses prepared statements
- `sslmode=require` is mandatory for Supabase
- Pool is configured with max 10 connections (Supabase free tier has ~60 total)

## Environment Variables

See `.env.example` — the Makefile loads `.env` automatically via `include`.

- `DATABASE_URL` — Supabase direct Postgres connection string (required, no default)
- `PORT` — Server port (default: 8080)
- `CORS_ORIGIN` — Allowed origin (default: http://localhost:5173)

## Guest Fields

`id` (UUID), `nome`, `sobrenome`, `telefone` (required strings), `relacionamento` ("noivo" | "noiva"), `confirmacao` (bool, default false), `created_at`, `updated_at`

## Conventions

- Domain-organized packages (all guest code in `internal/guest/`)
- Validation lives in the service layer, not handlers
- Handlers return JSON errors as `{"error": "message"}`
- Empty lists return `[]`, never `null`
- Import endpoint returns `{"imported": N, "errors": [...], "total": N}`

# Backend - ParaSempre

Wedding guest management API built in Go.

## Tech Stack

- **Go 1.26** with stdlib `net/http` (enhanced ServeMux with method routing)
- **pgx/v5** for PostgreSQL (Supabase) — no ORM
- **excelize/v2** for XLSX import, stdlib `encoding/csv` for CSV
- **log/slog** for structured logging

## Architecture

Three-layer domain-oriented structure: **Handler -> Service -> Repository**.

```
backend/
  cmd/server/main.go          # Entrypoint, CORS middleware, graceful shutdown, seed
  internal/
    config/config.go           # Env-based config (DB, PORT, CORS, Couple)
    database/database.go       # pgxpool connection helper
    guest/                     # Guest domain package
      model.go                 # Guest, CreateGuestInput, UpdateGuestInput
      repository.go            # Repository interface (incl. GetByPhone)
      repository_postgres.go   # pgx implementation
      service.go               # Business logic + validation
      handler.go               # HTTP handlers + route registration
      importer.go              # CSV/XLSX parsing (io.Reader based)
      *_test.go                # Tests for service, handler, importer
    user/                      # User domain package
      model.go                 # User, RegisterInput, CheckResponse
      repository.go            # Repository interface
      repository_postgres.go   # pgx implementation
      service.go               # Register, CheckByPhone, SeedCouple
      handler.go               # HTTP handlers + route registration
      *_test.go                # Tests for service, handler
  migrations/
    001_create_guests.sql
    002_create_users.sql       # users table
```

## Key Patterns

- **Repository is an interface** — tests use a mock with function fields (no codegen)
- **Pointer fields** in `UpdateGuestInput` (`*string`, `*bool`) for partial updates via `COALESCE`
- **Parsers accept `io.Reader`** — no file path coupling, fully testable
- **Table-driven tests** with `testing` stdlib only (no testify)
- **CORS middleware** is a single function in `main.go`
- **BIGINT GENERATED ALWAYS AS IDENTITY** primary keys (PostgreSQL best practice)
- **Cross-domain injection** — user.Service receives guest.Repository for phone lookup
- **Couple seed** — `SeedCouple()` called on startup, idempotent (skips if URACF exists)

## Commands

```bash
make test    # Run all unit tests
make run     # Start server (loads .env automatically)
make build   # Build binary to bin/server
make migrate # Run SQL migrations against Supabase
```

## API Endpoints

| Method | Path                 | Description                      |
|--------|----------------------|----------------------------------|
| GET    | /api/guests          | List all guests                  |
| POST   | /api/guests          | Create guest                     |
| GET    | /api/guests/{id}     | Get guest by ID                  |
| PUT    | /api/guests/{id}     | Update guest                     |
| DELETE | /api/guests/{id}     | Delete guest                     |
| POST   | /api/guests/import   | Import CSV/XLSX                  |
| POST   | /api/users           | Register user (phone+uracf)     |
| GET    | /api/users/check     | Check if user exists (?phone=)   |

## Supabase Connection

We connect **directly to PostgreSQL via pgx** — no Supabase REST API, no API keys needed.

- Use the **direct connection** string from Supabase Dashboard (Settings > Database > Connection string > URI)
- Must use port **5432** (direct), NOT 6543 (transaction pooler) — pgx uses prepared statements
- `sslmode=require` is mandatory for Supabase
- Pool is configured with max 10 connections (Supabase free tier has ~60 total)

## Environment Variables

See `.env.example` — the Makefile loads `.env` automatically via `include`.

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` — Supabase connection
- `PORT` — Server port (default: 8080)
- `CORS_ORIGIN` — Allowed origin (default: http://localhost:3000)
- `GROOM_FIRST_NAME`, `GROOM_LAST_NAME`, `GROOM_PHONE`, `GROOM_URACF` — Groom seed data
- `BRIDE_FIRST_NAME`, `BRIDE_LAST_NAME`, `BRIDE_PHONE`, `BRIDE_URACF` — Bride seed data

## Guest Fields

`id` (BIGINT), `first_name`, `last_name`, `phone` (TEXT nullable, UNIQUE), `relationship` ("P" | "R"), `confirmed` (bool, default false), `family_group` (BIGINT), `created_by` (TEXT, RACF 5 chars), `updated_by` (TEXT, RACF 5 chars), `created_at`, `updated_at`

## User Fields

`id` (BIGINT), `guest_id` (BIGINT nullable FK -> guests), `role` ("guest" | "groom" | "bride"), `uracf` (TEXT UNIQUE, 5 uppercase alphanumeric chars), `created_at`, `updated_at`

## Domain Validation Rules

- **Phone (BR mobile)**: `^\d{2}9\d{8}$` — 11 digits: DDD + 9 + 8 digits (e.g. 11999999999)
- **URACF**: `^[A-Z0-9]{5}$` — exactly 5 uppercase alphanumeric characters (e.g. TST01)
- **Relationship**: must be "P" (noivo/a side) or "R" (noiva/o side)

## Authentication

- Guest endpoints (POST, PUT, DELETE) require the `user-racf` HTTP header for audit trail (`created_by`/`updated_by`)
- User register endpoint is public (self-service: phone + uracf)
- User check endpoint is public (check by phone)
- System is **passwordless** — future OTP via WhatsApp for actual auth
- Groom/bride users are seeded on startup via URACF env vars (no guest_id, role = "groom"/"bride")

## Conventions

- Domain-organized packages (all guest code in `internal/guest/`, user in `internal/user/`)
- Validation lives in the service layer, not handlers
- Handlers return JSON errors as `{"error": "message"}`
- Empty lists return `[]`, never `null`
- Import endpoint returns `{"imported": N, "errors": [...], "total": N}`
- `GetByPhone` / `GetByURACF` / `GetByGuestID` return `(nil, nil)` when not found (not an error)

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Room booking service (Avito internship test task) — Go REST API for managing meeting rooms, schedules, time slots, and bookings with JWT auth. PostgreSQL backend.

## Build & Run

```bash
# Full stack (app + postgres) via Docker
docker-compose up --build

# Run locally (requires DB_URL and JWT_SECRET env vars, or .env file)
go build -o server ./cmd/server/main.go
./server

# Run all tests
go test ./...

# Run a single test
go test ./internal/service -run TestCreateBooking

# Test coverage
go test -coverprofile=cover.out ./...
go tool cover -func=cover.out

# Regenerate OpenAPI server code (requires oapi-codegen installed)
go generate ./internal/api/
```

## Architecture

Three-layer architecture: **handler → service → repository**, with domain models shared across layers.

- **`api.yaml`** — OpenAPI 3.0 spec, the source of truth for all endpoints
- **`internal/api/`** — Auto-generated code from `api.yaml` via `oapi-codegen`. Generates chi server interface (`ServerInterface`), request/response models, and embedded spec. **Do not edit `api.gen.go` by hand** — regenerate with `go generate ./internal/api/`
- **`internal/domain/`** — Domain models (`Room`, `Schedule`, `Slot`, `Booking`, `User`) and sentinel errors. These are the internal types used across all layers
- **`internal/handler/`** — HTTP handlers. `Server` struct implements the generated `api.ServerInterface`. Includes auth middleware that extracts JWT claims into context (`UserIDKey`, `RoleKey`). Error responses go through `writeError()`, success through `respondJSON()`
- **`internal/service/`** — Business logic layer. Each service accepts repository interfaces (defined in `interfaces.go`), enabling mock-based unit testing
- **`internal/repository/`** — PostgreSQL implementations using `pgx/v5` connection pool. `DB.WithTx()` provides transaction support
- **`cmd/server/main.go`** — Wiring: creates repos → services → handler, registers chi routes via `api.HandlerWithOptions` with auth middleware
- **`migrations/init.sql`** — Database schema, auto-applied by Postgres container on first start. Includes seed data for fixed admin/user UUIDs

## Key Design Decisions

- Slots are materialized in the DB (not computed on-the-fly) — generated from schedule when slots are requested, upserted with `ON CONFLICT DO NOTHING`
- Fixed UUIDs for dummy auth: admin=`11111111-...`, user=`22222222-...`
- Schedule is immutable — one schedule per room, cannot be changed after creation
- Booking cancellation is idempotent
- All times in UTC
- oapi-codegen middleware handles request validation and parameter parsing; the `OapiErrorHandler` in `handler/error.go` converts parsing failures to structured error responses

## Environment Variables

- `DB_URL` — Postgres connection string (required)
- `JWT_SECRET` — JWT signing key (required)
- `SERVER_PORT` — defaults to `8080`

## Testing

Unit tests use hand-written mocks in `internal/service/mock_test.go` implementing the repo interfaces. Handler tests in `internal/handler/handler_test.go` are integration/E2E style hitting the actual HTTP endpoints. CI runs e2e tests via GitHub Actions against the Docker Compose stack.
# Estate Access Control MVP

Backend-only MVP for a multi-tenant, estate-scoped RBAC system in Go. The implementation is intentionally centered on authorization correctness, active request context, permission drift, and trust boundaries instead of CRUD breadth.

## Architecture

Request flow:

`handler -> service/usecase -> authorizer -> repository -> database`

Handlers only parse input, validate shape, call services, and return responses. Authorization decisions never live inside HTTP handlers. This keeps trust-sensitive logic centralized, testable, and reusable across delivery layers.

Core modules:

- `cmd/server`: process entrypoint and graceful shutdown.
- `internal/app`: dependency injection and route wiring.
- `internal/config`: environment-driven config and JSON logging setup.
- `internal/database`: SQLite connection, migrations, and seed data.
- `internal/entities`: GORM models with UUID v7 identifiers.
- `internal/repositories`: data access boundaries.
- `internal/authorization`: fresh DB-backed permission resolution.
- `internal/services`: login, user, gate, and debug use cases.
- `internal/handlers`: thin Fiber handlers and centralized error handling.
- `internal/middleware`: request ID, auth, and structured request logging.
- `internal/utils`: Argon2 hashing, JWT utilities, validation, and app errors.
- `tests`: integration tests for drift and override behavior.

## Why These Choices

- Fiber: small, fast, ergonomic, and productive for focused API work without forcing framework-heavy patterns.
- GORM: practical migrations, transactions, joins, and repository-friendly persistence for an MVP that still wants production-style discipline.
- SQLite: zero-ops local persistence, deterministic test setup, and enough relational integrity for the challenge.
- Argon2id: modern password hashing resistant to GPU cracking and appropriate for credential storage.
- UUID v7: sortable identifiers with better index locality and production-friendly chronology than random UUID v4.

## Authorization Design

JWTs contain identity and session metadata only. They do not contain roles, permissions, or estate access. The backend always resolves authorization from the database on every request using:

- authenticated user ID from JWT
- active estate from `X-Estate-ID`
- fresh membership lookup
- fresh override lookup
- fresh estate-scoped role permission lookup

Resolution order:

1. Explicit `DENY` override
2. Explicit `ALLOW` override
3. Estate role permission
4. Default deny

This matters because of permission drift. If a user logs in while they are an Admin and later gets downgraded to Resident, the same JWT must stop working immediately for `POST /gate/open`. Trusting JWT-embedded roles or cached permissions would create stale authorization and silently over-authorize users.

Redis was intentionally excluded from authorization state because this challenge prioritizes correctness under changing permissions. A cache adds invalidation risk and can mask the exact bug being exercised here.

## Run

Requirements:

- Go 1.26+

Environment variables:

- `SERVER_PORT` default: `:8080`
- `DATABASE_DSN` default: `app.db`
- `JWT_SECRET` default: `change-me-in-production`
- `DEFAULT_USER_EMAIL` default: `admin@jarakey.com`
- `DEFAULT_USER_PASSWORD` default: `Pa$$w0rd!`

Start the server:

```bash
go run ./cmd/server
```

Run tests:

```bash
go test ./...
```

## Seeded Defaults

- User email: `admin@jarakey.com`
- User password: `Pa$$w0rd!`
- Estate: `Maple Residency`
- Roles: `admin`, `resident`
- Permission: `gate.open`

## Example Requests

Login:

```bash
curl -s http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@jarakey.com","password":"Pa$$w0rd!"}'
```

Get current user:

```bash
curl -s http://localhost:8080/me \
  -H "Authorization: Bearer $TOKEN"
```

Open gate:

```bash
curl -s -X POST http://localhost:8080/gate/open \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Estate-ID: $ESTATE_ID"
```

Simulate permission drift:

```bash
curl -s -X POST http://localhost:8080/debug/downgrade-role \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Estate-ID: $ESTATE_ID"
```

Then retry `POST /gate/open` with the same token. It should return `403`.

## Theory Answers

1. Minimum data required to authorize a request: authenticated user ID, active estate ID, requested permission code, and fresh database state for memberships, overrides, and estate-scoped role permissions.
2. Authorization logic should live in a dedicated authorization or service layer, not handlers, so trust-sensitive rules stay centralized, testable, and reusable.
3. The most dangerous RBAC bug here is stale over-authorization, where a user keeps access after a downgrade because the backend trusted cached or token-embedded permissions.
4. The backend should never trust client-supplied roles, permissions, estate access, or any authorization claims beyond an authenticated identity token that is still verified server-side.
5. AI helped accelerate scaffolding, test case coverage, and documentation drafting.
6. Guardrails were applied by keeping authorization decisions in a dedicated layer, validating every request against the database, and writing tests for permission drift and overrides.
7. The most important test written first is the same-JWT permission drift test, because it proves the core requirement that authorization is always re-evaluated fresh.

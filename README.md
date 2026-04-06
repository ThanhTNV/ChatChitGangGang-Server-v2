[![ci](https://github.com/ThanhTNV/ChatChitGangGang-Server-v2/actions/workflows/ci.yml/badge.svg)](https://github.com/ThanhTNV/ChatChitGangGang-Server-v2/actions/workflows/ci.yml)
[![release](https://github.com/ThanhTNV/ChatChitGangGang-Server-v2/actions/workflows/release.yml/badge.svg)](https://github.com/ThanhTNV/ChatChitGangGang-Server-v2/actions/workflows/release.yml)
# Internal Comm Backend

Go service for an internal communications backend: HTTP API with optional **Keycloak (OIDC/JWT)** auth, **PostgreSQL** for users/channels/messages, and a **WebSocket** entry point (real-time features are still evolving). Structured logging uses **`slog`**; routing uses **[chi](https://github.com/go-chi/chi)**.

**Module:** `github.com/chatchitganggang/internal-comm-backend`  
**Go:** 1.25+

## Features

- `GET /health` — liveness  
- `GET /ready` — readiness (optional Postgres ping when `DATABASE_URL` is set)  
- `GET /openapi.yaml`, `GET /docs` — OpenAPI 3 + Swagger UI  
- With **`KEYCLOAK_ISSUER`** + **`DATABASE_URL`**: `GET /v1/me`, `GET /v1/channels`, `GET /v1/ws` (JWT validation, user upsert)

If `KEYCLOAK_ISSUER` is unset, **`/v1/*` is not mounted** (only health, ready, and docs).

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)  
- [Docker Compose](https://docs.docker.com/compose/) (recommended for Postgres, Keycloak, Redis, MinIO locally)  
- Optional: `golangci-lint` for the same checks as CI

## Quick start (API on the host)

1. Clone the repo and copy env template:

   ```bash
   cp .env.example .env
   ```

2. Start Postgres (creates `internal_comm` via init script):

   ```bash
   docker compose up -d postgres
   ```

3. Apply migrations:

   ```bash
   make migrate-up
   ```

   On Windows without GNU Make: `make.cmd migrate-up` or `go run ./cmd/migrate -direction up` (load `DATABASE_URL` from your environment).

4. Run the server:

   ```bash
   make run
   ```

5. Open [http://localhost:8080/docs](http://localhost:8080/docs) (or your `HTTP_ADDR`).

To exercise **`/v1/*`**, set `KEYCLOAK_ISSUER`, `KEYCLOAK_AUDIENCE`, and `DATABASE_URL` in `.env`, run Keycloak (e.g. `docker compose up -d keycloak`), configure a realm and client, then restart the API. See [docs/keycloak-flutter.md](docs/keycloak-flutter.md) and [docs/keycloak-browser.md](docs/keycloak-browser.md).

## Run API inside Docker Compose

Optional profile **`app`** builds and runs the API in Compose (see `docker-compose.yml` for `KEYCLOAK_*` and `DATABASE_URL` tuned for service DNS):

```bash
docker compose --profile app up -d --build
```

## Environment variables

See **[`.env.example`](.env.example)** for the full list and comments. The process reads:

| Variable | Role |
|----------|------|
| `HTTP_ADDR` | Listen address (default `:8080`) |
| `LOG_LEVEL`, `LOG_FORMAT` | `slog` level and `text` / `json` |
| `DATABASE_URL` | Postgres DSN; enables `/ready` DB check and is **required** when Keycloak is enabled |
| `KEYCLOAK_ISSUER` | JWT `iss` (must match token); if empty, `/v1` routes are disabled |
| `KEYCLOAK_AUDIENCE` | Required with issuer; must match token `aud` or `azp` |
| `KEYCLOAK_JWKS_URL` | Optional; defaults to `{issuer}/protocol/openid-connect/certs` |

`REDIS_*` and `MINIO_*` in `.env.example` are reserved for future features; the server does not read them yet.

## Makefile

| Target | Description |
|--------|-------------|
| `make test` | `go test ./...` |
| `make test-race` | Race detector (needs CGO; CI uses this on Linux) |
| `make lint` | `golangci-lint run` |
| `make run` | `go run ./cmd/server` |
| `make migrate-up` / `migrate-down` | Embedded SQL migrations |

Windows: use **`make.cmd`** or **`scripts/ps/dev.ps1`** if GNU Make is not installed.

## Project layout

```
cmd/server/       # HTTP server entrypoint
cmd/migrate/      # DB migrations CLI
internal/auth/    # JWKS, JWT validation, bearer middleware
internal/channel/ # Channel repository
internal/config/  # Env loading
internal/dbmigrate/ # Embedded migrations
internal/httpserver/ # Routes, OpenAPI/Swagger UI, WebSocket stub
internal/user/    # User repository (Keycloak sub upsert)
docs/             # Keycloak + client notes
output/           # Implementation plan and dated progress specs
```

## Documentation

- [docs/keycloak-flutter.md](docs/keycloak-flutter.md) — mobile / public client tokens, `aud` vs `azp`  
- [docs/keycloak-browser.md](docs/keycloak-browser.md) — browser sessions and OIDC flow  
- [output/backend-golang-implementation-plan.md](output/backend-golang-implementation-plan.md) — roadmap checklists  

## CI

GitHub Actions (`.github/workflows/ci.yml`) runs `go test -race ./...` and `golangci-lint` on push/PR.

## License

Add a `LICENSE` file if you intend to open-source or distribute this repository.

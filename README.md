[![ci](https://github.com/ThanhTNV/ChatChitGangGang-Server-v2/actions/workflows/ci.yml/badge.svg)](https://github.com/ThanhTNV/ChatChitGangGang-Server-v2/actions/workflows/ci.yml)
[![release](https://github.com/ThanhTNV/ChatChitGangGang-Server-v2/actions/workflows/release.yml/badge.svg)](https://github.com/ThanhTNV/ChatChitGangGang-Server-v2/actions/workflows/release.yml)
# Internal Comm Backend

Go service for an internal communications backend: HTTP API with optional **Keycloak (OIDC/JWT)** auth, **PostgreSQL** for users/channels/messages, and a **WebSocket** entry point (real-time features are still evolving). Structured logging uses **`slog`**; routing uses **[chi](https://github.com/go-chi/chi)**.

**Module:** `github.com/chatchitganggang/internal-comm-backend`  
**Go:** 1.25+

### Docker Compose files

| File | Purpose |
|------|---------|
| **`docker-compose.yml`** | Default stack: Postgres, Redis, MinIO, Keycloak (**`http://localhost:8090`** — `/realms/...`, `/admin/...`), optional **`api`** (`--profile app`) on **`http://localhost:8080`**. **No** TLS reverse proxy. |
| **`docker-compose.ingress.yml`** | **Overlay** — merge with the file above. Adds **nginx** on **80/443**, **`https://localhost/api`**, **`https://localhost/oauth`**. Sets Keycloak **`KC_HTTP_RELATIVE_PATH=/oauth`**, updates API **`KEYCLOAK_ISSUER`** / **`KEYCLOAK_JWKS_URL`**, **`OPENAPI_PUBLIC_BASE_URL`**. Requires TLS cert paths inside that file. |

Nginx config for ingress: **`deploy/docker/nginx/nginx.conf`** (Keycloak **`proxy_pass`** keeps the **`/oauth`** prefix).

## Features

- `GET /health` — liveness  
- `GET /ready` — readiness probes for **Postgres** (`DATABASE_URL`), **Redis** (`REDIS_ADDR`), **Keycloak** (HTTP GET `KEYCLOAK_READY_URL` or else `KEYCLOAK_ISSUER` when set), **MinIO** (`MINIO_ENDPOINT` + `MINIO_USE_SSL` → `/minio/health/live`); unset vars report `skipped` for that check  
- `GET /openapi.yaml`, `GET /docs` — OpenAPI 3 + Swagger UI  
- With **`KEYCLOAK_ISSUER`** + **`DATABASE_URL`**: `GET /v1/me`, `GET`/`POST /v1/channels`, `GET /v1/channels/{id}/messages`, `GET /v1/ws` (JWT validation, user upsert)

If `KEYCLOAK_ISSUER` is unset, **`/v1/*` is not mounted** (only health, ready, and docs).

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)  
- [Docker Compose](https://docs.docker.com/compose/) (recommended for Postgres, Keycloak, Redis, MinIO locally)  
- Optional: `golangci-lint` for the same checks as CI  

Set **`KEYCLOAK_ISSUER`** (and discovery URLs in your SPA) to match **direct** vs **ingress** Keycloak — see [docs/keycloak-browser.md](docs/keycloak-browser.md).

---

## Running the stack

Pick **one** primary path: [API on the host (Go)](#option-1-dependencies-in-docker-compose-api-with-go-on-the-host) or [API inside Compose](#option-2-full-stack-in-docker-compose-including-api). TLS ingress is a **second Compose file** merged with the first (see [TLS ingress](#tls-ingress-second-compose-file)).

### Option 1: Dependencies in Docker Compose, API with Go on the host

Use this when you want `go run`, debugging in your IDE, or hot reload while databases and Keycloak run in containers.

#### 1. Environment file

```bash
cp .env.example .env
```

Edit `.env` so the API reaches **published localhost ports** (not Docker service names):

| Setting | Typical value (host API) |
|--------|---------------------------|
| `DATABASE_URL` | `postgres://postgres:postgres@127.0.0.1:5432/internal_comm?sslmode=disable` |
| `REDIS_ADDR` | `localhost:6379` |
| `MINIO_ENDPOINT` | `localhost:9000` |
| `MINIO_USE_SSL` | `false` |
| `KEYCLOAK_ISSUER` | **Direct Keycloak:** `http://localhost:8090/realms/master`. **SPA via ingress:** `https://localhost/oauth/realms/master`. Must match JWT `iss`. |
| `KEYCLOAK_AUDIENCE` | Your client id (e.g. `internal-comm-api`) |
| `KEYCLOAK_JWKS_URL` | Optional; defaults to `{issuer}/protocol/openid-connect/certs`. **Merged ingress** sets an internal **`http://keycloak:8080/oauth/realms/.../certs`** URL on the **`api`** service (see **`docker-compose.ingress.yml`**). |
| `KEYCLOAK_READY_URL` | Optional. Compose **`api`** uses `http://keycloak:9000/health/ready`. For host API, unset and `/ready` probes `KEYCLOAK_ISSUER`, or point at a reachable URL on **8090**. |

Unset `KEYCLOAK_ISSUER` if you only want health/openapi without `/v1`.

#### 2. Start backing services

Minimal (Postgres only):

```bash
docker compose up -d postgres
```

Full local stack used by the API (Postgres, Redis, MinIO, Keycloak):

```bash
docker compose up -d postgres redis minio keycloak
```

Wait until Postgres is healthy (`docker compose ps`). Postgres listens on **`127.0.0.1:5432`** on the host.

#### 3. Migrations

```bash
make migrate-up
```

Windows without GNU Make: `make.cmd migrate-up` or:

```bash
go run ./cmd/migrate -direction up
```

(Ensure `DATABASE_URL` is set in the environment or loaded from `.env` the same way you run the server.)

#### 4. Keycloak (only if you enabled `KEYCLOAK_ISSUER`)

1. Open **http://localhost:8090/admin/** when using **`docker-compose.yml` only**. With **ingress** merged in, use **`https://localhost/oauth/admin/...`** (published **8090** then uses **`/oauth/...`** paths — see keycloak-browser doc).
2. Create or select a **realm** (e.g. `master` matches the example issuer).
3. Create an OIDC **client** whose **audience** / **azp** matches `KEYCLOAK_AUDIENCE`.
4. Enable **Direct access grants** if you use password grant (e.g. Postman).

Details: [docs/keycloak-flutter.md](docs/keycloak-flutter.md), [docs/keycloak-browser.md](docs/keycloak-browser.md).

#### 5. Run the API

```bash
make run
```

or:

```bash
go run ./cmd/server
```

#### 6. Verify

- Swagger: [http://localhost:8080/docs](http://localhost:8080/docs) (or your `HTTP_ADDR`)
- Liveness: `GET http://localhost:8080/health`
- Readiness: `GET http://localhost:8080/ready`

**Note:** With **ingress**, the public API base is **`https://localhost/api`** (not your host `go run` port). For a host-run API behind nginx you would point nginx at `host.docker.internal:8080` (or your OS equivalent).

---

### Option 2: Full stack in Docker Compose (including API)

Use this when you want the Go service built and run as a container with in-cluster DNS (`postgres`, `keycloak`, etc.).

#### 1. Prerequisites

- Same repo checkout; no requirement for a host `.env` unless you override Compose with `env_file` (by default the **`api`** service env is defined in `docker-compose.yml`).

#### 2. Build and start API + dependencies

```bash
docker compose --profile app up -d --build
```

This starts **postgres**, **redis**, **minio**, **keycloak**, and **api** (profile `app`). The API image is built from the repo `Dockerfile`.

#### 3. Migrations

Run once against the same Postgres instance (from the host, using published Postgres):

```bash
# Example if DATABASE_URL matches compose credentials and 127.0.0.1:5432
set DATABASE_URL=postgres://postgres:postgres@127.0.0.1:5432/internal_comm?sslmode=disable
go run ./cmd/migrate -direction up
```

Or run a one-off migrate container if you add one; today the repo expects **`make migrate-up`** from the host against `127.0.0.1:5432`.

#### 4. Verify

- API (direct to container publish): **http://localhost:8080** — e.g. `/docs`, `/health`, `/ready`, `/v1/*` with a valid JWT.
- Keycloak (direct HTTP): **http://localhost:8090/** (e.g. `/realms/master/...`, admin **`/admin/`**).

The **`api`** service in **`docker-compose.yml`** uses **`KEYCLOAK_ISSUER=http://localhost:8090/realms/master`**. When you merge **`docker-compose.ingress.yml`**, that overlay switches issuer and JWKS to **`https://localhost/oauth/...`**.

---

### TLS ingress (second Compose file)

Use **`docker-compose.ingress.yml`** **with** **`docker-compose.yml`** (do not run the ingress file alone).

- **`https://<host>/api/...`** → **`api:8080`** with the **`/api`** prefix stripped.
- **`https://<host>/oauth/...`** → **`keycloak:8080/oauth/...`** with **`/oauth` preserved**. The overlay sets **`KC_HTTP_RELATIVE_PATH=/oauth`** and **`KC_HOSTNAME=https://localhost/oauth`** so TLS termination at nginx still produces correct admin-console and OIDC URLs.

Cert paths: **`docker-compose.ingress.yml`** (defaults `C:/Users/Admin/cert.pem`, `key.pem` — edit for your machine). Use **`--profile app`** so **`api`** and **`ingress`** start (the ingress service is declared with **`profiles: ["app"]`** in the overlay so Compose stays valid).

```bash
docker compose -f docker-compose.yml -f docker-compose.ingress.yml --profile app up -d --build
```

**https://localhost/api/docs**, **https://localhost/oauth/** (redirect to **`/oauth/admin/...`** is normal). WebSocket: **`wss://localhost/api/v1/ws?access_token=...`**.

**Troubleshooting:** Clear site data for **`https://localhost`** after changing Keycloak or nginx. Admin UI stuck loading usually meant nginx and Keycloak disagreed on **`/oauth`**; this split uses **matching** relative path + **`proxy_pass .../oauth/`**. **`502`:** wait for Keycloak health; check **`docker compose logs keycloak`**. Debug: **`docker compose exec ingress wget -qO- http://keycloak:8080/oauth/realms/master/.well-known/openid-configuration`**.

**Switching** direct ↔ ingress with the same **`keycloak_data`** volume can leave bad hostname metadata; for dev you can drop that volume.

**Profiles**

| Profile | Where | What it adds |
|--------|--------|----------------|
| *(default)* | base compose | `postgres`, `redis`, `minio`, `keycloak` |
| `app` | base compose | **`api`** (Go HTTP on **8080**) |
| `app` | **`docker-compose.ingress.yml`** | **`ingress`** (nginx **80/443**) — only when both files are merged |
| `livekit` | base compose | LiveKit SFU (optional) |

---

## Quick reference (copy-paste)

**Host Go API + all deps**

```bash
cp .env.example .env
# Edit .env: DATABASE_URL, KEYCLOAK_*, etc.
docker compose up -d postgres redis minio keycloak
make migrate-up
make run
```

**Everything in Docker (API container)**

```bash
docker compose --profile app up -d --build
# Run migrations from host against 127.0.0.1:5432 (see Option 2)
```

**Docker API + TLS ingress**

```bash
docker compose -f docker-compose.yml -f docker-compose.ingress.yml --profile app up -d --build
```

## Environment variables

See **[`.env.example`](.env.example)** for the full list and comments. The process reads:

| Variable | Role |
|----------|------|
| `HTTP_ADDR` | Listen address (default `:8080`) |
| `LOG_LEVEL`, `LOG_FORMAT` | `slog` level and `text` / `json` |
| `DATABASE_URL` | Postgres DSN; `/ready` DB ping when set; **required** when Keycloak is enabled |
| `REDIS_ADDR` | Redis address for `/ready` (e.g. `localhost:6379`); optional |
| `MINIO_ENDPOINT` | Host (or `host:port`) for `/ready` MinIO liveness; optional |
| `MINIO_USE_SSL` | `true` / `false` — scheme for MinIO health URL |
| `KEYCLOAK_ISSUER` | JWT `iss`; required for `/v1` when set; if empty, `/v1` disabled and Keycloak check skipped |
| `KEYCLOAK_READY_URL` | Optional; URL for `/ready` Keycloak GET. Keycloak 26 serves `/health/ready` on **port 9000** (management) by default — use e.g. `http://keycloak:9000/health/ready` in Compose when `KEYCLOAK_ISSUER` is a host URL |
| `KEYCLOAK_AUDIENCE` | Required with issuer; must match token `aud` or `azp` |
| `KEYCLOAK_JWKS_URL` | Optional; defaults to `{issuer}/protocol/openid-connect/certs` |
| `OPENAPI_PUBLIC_BASE_URL` | Optional; absolute API base for OpenAPI “servers” + Swagger “Try it out” (no trailing slash). **Ingress overlay** sets **`https://localhost/api`**. If unset on the server, derived from `Host` and `X-Forwarded-*` (nginx sends **`X-Forwarded-Prefix: /api`** for the API location). |

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
cmd/server/            # HTTP server entrypoint
cmd/migrate/           # DB migrations CLI
deploy/docker/nginx/   # TLS ingress (used with docker-compose.ingress.yml)
internal/apiembed/     # Embedded OpenAPI YAML
internal/auth/         # JWKS, JWT validation, bearer middleware
internal/channel/      # Channels store + repository
internal/chat/         # Messages store, cursor pagination, repository
internal/config/       # Env loading
internal/dbmigrate/    # Embedded SQL migrations
internal/httpserver/   # Routes, health/ready, OpenAPI/Swagger, channels, messages, WebSocket
internal/user/         # User sync / repository (Keycloak `sub` upsert)
docs/                  # Keycloak + client notes
output/                # Implementation plan, API integration guides, Postman packs, progress specs
```

## Documentation

- [docs/keycloak-flutter.md](docs/keycloak-flutter.md) — mobile / public client tokens, `aud` vs `azp`  
- [docs/keycloak-browser.md](docs/keycloak-browser.md) — browser sessions, OIDC discovery (direct vs ingress), CORS  
- [output/api-guide/](output/api-guide/) — dated client integration guides (web/mobile vs direct vs TLS ingress)  
- [output/backend-golang-implementation-plan.md](output/backend-golang-implementation-plan.md) — roadmap checklists  

## CI

GitHub Actions (`.github/workflows/ci.yml`) runs `go test -race ./...` and `golangci-lint` on push/PR.

## License

Add a `LICENSE` file if you intend to open-source or distribute this repository.

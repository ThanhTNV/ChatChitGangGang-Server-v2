[![ci](https://github.com/ThanhTNV/ChatChitGangGang-Server-v2/actions/workflows/ci.yml/badge.svg)](https://github.com/ThanhTNV/ChatChitGangGang-Server-v2/actions/workflows/ci.yml)
[![release](https://github.com/ThanhTNV/ChatChitGangGang-Server-v2/actions/workflows/release.yml/badge.svg)](https://github.com/ThanhTNV/ChatChitGangGang-Server-v2/actions/workflows/release.yml)
# Internal Comm Backend

Go service for an internal communications backend: HTTP API with optional **Keycloak (OIDC/JWT)** auth, **PostgreSQL** for users/channels/messages, and a **WebSocket** entry point (real-time features are still evolving). Structured logging uses **`slog`**; routing uses **[chi](https://github.com/go-chi/chi)**.

**Module:** `github.com/chatchitganggang/internal-comm-backend`  
**Go:** 1.25+

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

Keycloak listens on **`http://localhost:8090`** with **standard** OIDC paths (`/realms/...`). Optional TLS **ingress** maps **`https://<host>/oauth/...`** to the same Keycloak (nginx strips `/oauth` and sets **`X-Forwarded-Prefix`**). Set **`KEYCLOAK_ISSUER`** on the API to match how your clients reach Keycloak (see table in [docs/keycloak-browser.md](docs/keycloak-browser.md)).

---

## Running the stack

Pick **one** primary path: [API on the host (Go)](#option-1-dependencies-in-docker-compose-api-with-go-on-the-host) or [API inside Compose](#option-2-full-stack-in-docker-compose-including-api). TLS ingress is optional and only wired to the **container** API (see [TLS ingress](#optional-tls-ingress-nginx)).

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
| `KEYCLOAK_JWKS_URL` | e.g. `http://localhost:8090/realms/master/protocol/openid-connect/certs` (or omit to derive from issuer) |
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

1. Open **http://localhost:8090/admin/** (or **https://localhost/oauth/admin/** when using ingress).
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

**Note:** The default **ingress** profile sends `/api` to the **`api` container**, not to a process on your machine. For a host-run API, call **`http://localhost:8080`** directly (no `/api` prefix). To put a host API behind nginx you would change `deploy/docker/nginx/nginx.conf` to proxy to `host.docker.internal:8080` (or your OS equivalent).

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

The **`api`** service defaults to **`KEYCLOAK_ISSUER=http://localhost:8090/realms/master`** (direct). If your SPA uses **only** **`https://localhost/oauth/...`**, change it to **`https://localhost/oauth/realms/master`** in `docker-compose.yml` (or override via env) so JWT `iss` matches.

---

### Optional: TLS ingress (nginx)

Terminates TLS on **443**; HTTP **80** redirects to HTTPS. Routes:

- **`https://<host>/api/...`** → Go API (**`api`** service, prefix `/api` stripped → backend paths like `/v1/...`, `/docs`, `/health`).
- **`https://<host>/oauth/...`** → Keycloak at **`keycloak:8080/`** ( **`/oauth`** stripped; standard **`/realms/...`** on the app port).

Requires TLS files on the host (paths in `docker-compose.yml` under **`ingress`**, e.g. `C:/Users/Admin/cert.pem` and `key.pem`). Profile: **`ingress`**.

```bash
docker compose --profile app --profile ingress up -d --build
```

Then use **https://localhost/api/docs** and **https://localhost/oauth/realms/...** (admin: **https://localhost/oauth/admin/**). WebSocket example: **`wss://localhost/api/v1/ws?access_token=...`**.

**404 on `https://localhost/admin/` (no `/oauth`):** Keycloak often responds with path-absolute URLs such as **`/admin`**, **`/realms/...`**, **`/resources/...`**. The browser turns those into **`https://localhost/admin`**, but only **`/oauth/...`** is proxied. Nginx redirects **`/admin`**, **`/realms/`**, and **`/resources/`** to the same path under **`/oauth/`**. Open the console at **`https://localhost/oauth/admin/`** and use **`https://localhost/oauth/...`** in SPA authority / discovery URLs.

If the admin console shows **ERR_TOO_MANY_REDIRECTS**: (1) **Do not** set **`KC_HTTP_RELATIVE_PATH`** on Keycloak while nginx **strips** **`/oauth`** (that mismatch makes Keycloak redirect forever). This repo’s compose keeps Keycloak at **`/`** and relies on **`X-Forwarded-Prefix`**. (2) Clear **cookies** for **`https://localhost`**. (3) Ensure **`KC_PROXY_HEADERS=xforwarded`** and **`KC_HTTP_ENABLED=true`**; if redirects are still wrong, uncomment **`KC_HOSTNAME=https://localhost`** (hostname only, no **`/oauth`** path). Nginx rewrites common **`http://…`** `Location` headers to **`https://$host/oauth/`**. After changing env: **`docker compose up -d --force-recreate keycloak`** and reload nginx if needed.

If nginx shows **502 Bad Gateway** on **`/oauth`** or **`/api`**: (1) Start **both** profiles so the **`api`** service exists: `docker compose --profile app --profile ingress up -d`. (2) Wait **1–2 minutes** after Keycloak starts. (3) Check **`docker compose logs keycloak`** and **`docker compose ps`**. (4) From the ingress container, **`wget -qO- http://keycloak:8080/realms/master/.well-known/openid-configuration --timeout=5`** should return JSON when Keycloak is up (backend path has **no** `/oauth` prefix).

**Profiles reference**

| Profile | What it adds |
|--------|----------------|
| *(default)* | `postgres`, `redis`, `minio`, `keycloak` |
| `app` | **`api`** (Go HTTP server) |
| `ingress` | **nginx** TLS reverse proxy (needs cert files on disk) |
| `livekit` | LiveKit SFU (optional; see `docker-compose.yml`) |

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
docker compose --profile app --profile ingress up -d --build
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
| `OPENAPI_PUBLIC_BASE_URL` | Optional; absolute API base for OpenAPI/Swagger (no trailing slash). If unset, derived from `Host` and `X-Forwarded-*` (set **`X-Forwarded-Prefix: /api`** on your ingress for path-prefixed routes). |

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

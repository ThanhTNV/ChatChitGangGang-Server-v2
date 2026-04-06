# Progress summary — 2026-04-06

## Source of truth

- Plan: `output/backend-golang-implementation-plan.md`

## Checklist status

| Section | Item | Plan | Status | Evidence |
|---------|------|------|--------|----------|
| A | Module + layout | [x] | Done | `go.mod`, `cmd/server`, `cmd/migrate`, `internal/*` |
| A | Logging + request ID | [x] | Done | `internal/httpserver/server.go` (`slog`, chi RequestID) |
| A | Env config | [x] | Done | `internal/config/config.go` |
| A | Dockerfile | [x] | Done | `Dockerfile` |
| A | Makefile / tasks | [x] | Done | `Makefile`, `make.cmd`, `scripts/ps/dev.ps1` |
| A | CI | [x] | Done | `.github/workflows/ci.yml`, `release.yml` |
| B | Migration tool | [x] | Done | `internal/dbmigrate/`, `cmd/migrate` |
| B | Tables + index | [x] | Done | `internal/dbmigrate/migrations/000001_init.up.sql` |
| B | Seed script (optional) | [ ] | Open (optional) | No `cmd/seed` / seeds dir |
| C | JWKS + cache | [x] | Done | `internal/auth/jwks.go`, `validator.go` |
| C | JWT REST middleware | [x] | Done | `internal/auth/middleware.go`, `GET /v1/me` |
| C | WS token path | [x] | Done | `internal/httpserver/ws.go` (`access_token`, subprotocol) |
| C | sub → users | [x] | Done | `internal/user/repository.go` `UpsertByKeycloakSub` |
| C | Flutter / Keycloak doc | [x] | Done | `docs/keycloak-flutter.md` |
| D | GET /v1/channels | [ ] | Open | No channel handlers in `internal/httpserver` |
| D | POST /v1/channels | [ ] | Open | — |
| D | GET messages cursor | [ ] | Open | — |
| D | Member-only authz | [ ] | Open | — (depends on D routes) |
| E | WS hub + Redis | [ ] | Open | WS exists as **stub** (`ws.go` read loop only); no Redis client, no pub/sub |
| E | send_message / join / leave | [ ] | Open | — |
| E | Graceful shutdown | [ ] | Open | HTTP shutdown present; WS hub drain N/A |
| F | All items | [ ] | Open | `internal/file/doc.go` placeholder only |
| G | All items | [ ] | Open | No LiveKit code |
| H | `/health` + `/ready` (DB + Redis) | [ ] | **Mismatch / partial** | `/health`, `/ready` implemented; DB ping when pool set; **no Redis** in `/ready` |
| H | Prometheus | [ ] | Open | — |
| H | OpenTelemetry | [ ] | Open | — |
| H | Helm | [ ] | Open | — |
| I | All items | [ ] | Open | — |
| J | OpenAPI / WS payloads | [x] | Done (REST + connect) | `internal/apiembed/openapi.yaml`, `/docs`, `/openapi.yaml`; WS **message** schemas still TBD with E |
| J | `.env` example | [x] | Done | `.env.example` |
| J | Runbook | [ ] | Open | — |

## Mismatches (plan vs repo)

1. **H.1:** Checklist text expects **DB + Redis** on `/ready`, but the server only reports `database` (ok / skipped / error). Either add a Redis ping to `/ready` and keep `[ ]` until done, or adjust the plan wording to reflect “DB now, Redis when E lands.”
2. **E vs current WS:** `GET /v1/ws` validates JWT and upgrades, but there is no channel protocol, hub, or Redis — section **E** should stay **Open** until those exist.

## Recently completed (since prior 2026-04-06 brief)

- **C** — Full Keycloak path: JWKS, bearer middleware, `/v1/me`, `/v1/ws` stub, user upsert.
- **J** — OpenAPI 3 embedded and served; `.env.example` aligned.

## Recommended next slice

- **Primary:** **D.1** — `GET /v1/channels` (membership-filtered for current user), plus OpenAPI and tests (see [next-step-spec.md](./next-step-spec.md)).
- **Defer:** B.4 dev seed unless fixtures are needed before channel UI; H Redis check until Redis is used by the app.

# Progress summary — 2026-04-10

## Source of truth

- Plan: `output/backend-golang-implementation-plan.md`

## Checklist status

| Section | Item | Plan | Status | Evidence |
|---------|------|------|--------|----------|
| A | All six items | [x] | Done | `go.mod`, `cmd/*`, `internal/*`, `Dockerfile`, `Makefile` / `make.cmd`, `.github/workflows/ci.yml`, `release.yml` |
| B | Migrations + tables + index | [x] | Done | `internal/dbmigrate/migrations/000001_init.up.sql`, `cmd/migrate` |
| B | Seed script (optional) | [ ] | Open | No `cmd/seed` |
| C | All five items | [x] | Done | `internal/auth/*`, `/v1/me`, `/v1/ws`, `internal/user`, `docs/keycloak-flutter.md` |
| D | GET /v1/channels | [x] | Done | `internal/channel/repository.go`, `internal/httpserver/channels.go` |
| D | POST /v1/channels | [ ] | Open | No create handler |
| D | GET …/messages | [ ] | Open | — |
| D | Member-only authz | [ ] | Partial | List filtered by membership only; no write/message paths yet |
| E | Hub + Redis + protocol | [ ] | Open | `ws.go` upgrades + read loop stub; no Redis, no JSON message protocol |
| F | All items | [ ] | Open | `internal/file/doc.go` |
| G | All items | [ ] | Open | — |
| H | `/health` + `/ready` (DB + Redis) | [ ] | **Mismatch / partial** | Endpoints exist; DB ping when pool set; **no Redis** in `/ready` |
| H | Prometheus / OTel / Helm | [ ] | Open | — |
| I | All items | [ ] | Open | — |
| J | OpenAPI + `.env` | [x] | Done | `internal/apiembed/openapi.yaml`, `/docs`, `.env.example`, `docs/keycloak-browser.md` |
| J | Runbook | [ ] | Open | — |

## Mismatches (plan vs repo)

1. **H.1:** Plan text mentions **DB + Redis** on `/ready`; implementation only checks **database**. Either extend `/ready` when Redis is integrated (section E) or relax plan wording.
2. **§6 Security (non–A–J):** JWT validation is implemented for mounted `/v1` routes; full security checklist (rate limits, CORS, etc.) still open — track under **I** / product decisions.

## Recently completed (since 2026-04-06 checkpoint)

- CI: **golangci-lint v2** + `golangci-lint-action@v9`, `.golangci.yml` **version: "2"**; `ws.go` **errcheck** on `conn.Close`.
- **Release:** GHCR image tag **lowercased** `github.repository` in `release.yml`.
- **Docs / repo hygiene:** `README.md`, `.gitignore`, `docs/keycloak-browser.md`.

## Recommended next slice

- **D.2** `POST /v1/channels` — see [next-step-spec.md](./next-step-spec.md).

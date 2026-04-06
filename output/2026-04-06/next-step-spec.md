# Next step implementation spec — 2026-04-06

## Metadata

- **Targets checklist:** **D. Chat REST** — item 1 (`GET /v1/channels`), foundation for D.4 (membership authz on reads)
- **Estimated scope:** 1–2 days
- **Out of scope:** `POST /v1/channels`, messages API, Redis, WS protocol, Prometheus, Helm, runbook

## Goal

Implement **`GET /v1/channels`** returning channels the authenticated user belongs to, using existing Postgres schema (`channels`, `channel_members`), JWT context from existing middleware, and document the route in OpenAPI.

## Prerequisites

- [ ] Postgres reachable with migrations applied (`cmd/migrate` or `make migrate-up`)
- [ ] `DATABASE_URL` + Keycloak envs set for local `/v1` testing (same as current `/v1/me` flow)
- [ ] Optional: a few rows in `channels` + `channel_members` for manual verification (no seed script required for this slice)

## Design notes

- Reuse **`auth` context** (`user_id` / internal UUID) already set by `BearerMiddleware` after upsert.
- Follow **`internal/user`** style: small repository type with `pgxpool`, no new ORM.
- JSON shape: keep it minimal — e.g. `{"channels":[{"id","name","type","created_at",...}]}`; align field names with plan and Flutter expectations in one pass.
- **Authorization:** for this endpoint, membership is the filter (only channels where `channel_members.user_id = current user`). Deeper policy (e.g. direct vs group rules) can wait until `POST /v1/channels`.

## Files to add or modify

| Path | Action |
|------|--------|
| `internal/channel/repository.go` (or `internal/chat/`) | **Create** — list channels for `user_id` |
| `internal/httpserver/channels.go` (or extend `v1.go`) | **Create** — `handleListChannels` |
| `internal/httpserver/server.go` | **Modify** — register `GET /v1/channels` under bearer group; pass DB/repo deps |
| `cmd/server/main.go` | **Modify** — construct channel repo when `dbPool != nil` and auth enabled |
| `internal/apiembed/openapi.yaml` | **Modify** — path `/v1/channels`, schema for list response, 401 |
| `internal/httpserver/*_test.go` | **Modify** — handler test with mocked repo or httptest + sql mock if you already have a pattern; otherwise integration test behind build tag — **minimum:** unit test on repository with testcontainers **or** documented manual test steps if DB test infra not ready |

## API / types

- **REST:** `GET /v1/channels` — `Authorization: Bearer <jwt>`
- **200:** JSON array (or wrapped list) of channel objects the user is a member of; stable sort (e.g. `created_at DESC` or `name`)
- **401:** existing JSON `{"error":"..."}` from middleware
- **500:** DB errors as JSON or consistent with other handlers
- **OpenAPI:** extend `internal/apiembed/openapi.yaml` (`components.schemas` for channel + list)

## Implementation steps

1. Add **`Channel` struct** and **`ListForUser(ctx, userID uuid.UUID) ([]Channel, error)`** (or equivalent) querying `channels` joined with `channel_members` where `user_id = $1`.
2. Wire repository in **`main.go`** only when database + auth are configured (same guard as today).
3. Add **`GET /v1/channels`** inside the `/v1` group that already uses `a.Bearer`, calling repo with `auth.UserID(r.Context())`.
4. Update **`openapi.yaml`** with the new operation and response schema.
5. Add **tests** — prefer testing the SQL with a real DB (Testcontainers) if the project adds that dependency; otherwise add **handler tests** with a small **fake repo** interface to keep CI green without Postgres.

## Tests

- `go test ./...`
- `go test -race ./...` on touched packages
- `go vet ./...`
- If golangci-lint is in CI, run it locally on changed files

## Definition of done

- [ ] `GET /v1/channels` returns only member channels for the JWT user
- [ ] `internal/apiembed/openapi.yaml` updated; `/docs` shows the operation
- [ ] Tests pass in CI
- [ ] Update `output/backend-golang-implementation-plan.md` — tick **D.1** when satisfied (and **D.4** partially if membership enforcement is clearly implemented for this route only)

## Risks / follow-ups

- Empty DB → empty list vs 404: prefer **200 + `[]`** unless product says otherwise.
- **`POST /v1/channels`** and **message history** will reuse the same repo patterns; keep queries index-friendly (`idx_messages_channel_created` already exists for later).

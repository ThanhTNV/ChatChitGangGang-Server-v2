# Next step implementation spec — 2026-04-10

> **Status:** Implemented in-repo (verified 2026-04-06). Regenerate via **check-implementation-progress** for a fresh next step (e.g. **D.3** messages).

## Metadata

- **Targets checklist:** **D. Chat REST** — item 2 (`POST /v1/channels`), supports later D.3–D.4
- **Estimated scope:** 1–2 days
- **Out of scope:** message history API, Redis, WS message protocol, admin roles beyond “authenticated creator”, direct-channel pairing rules (can defer to a follow-up)

## Goal

Add **`POST /v1/channels`** so an authenticated user can create a **group** channel (minimal policy: any authenticated user may create), insert `channels` + `channel_members` (creator as `member` or `admin` — pick one and document), return the created channel JSON.

## Prerequisites

- [x] Postgres + migrations applied
- [x] Keycloak envs + `DATABASE_URL` for local `/v1` testing
- [x] Policy: any authenticated user may create **group** channels only; **direct** rejected with 400

## Design notes

- Reuse **`auth.UserID`** from bearer middleware.
- **Transaction:** `INSERT channels` + `INSERT channel_members` in one transaction (rollback on failure).
- **Validation:** `name` non-empty, bounded length (e.g. ≤256 runes); `type` enum `group` for this slice (reject `direct` with 400 unless you implement pairing).
- **Idempotency:** not required for v1; optional client-supplied `Idempotency-Key` out of scope.
- **OpenAPI:** add operation + request/response schemas in `internal/apiembed/openapi.yaml`.

## Files to add or modify

| Path | Action |
|------|--------|
| `internal/channel/repository.go` (or new `tx.go`) | **Modify** — `CreateGroup(ctx, userID, name string) (Channel, error)` or similar |
| `internal/channel/channel.go` | **Modify** — request DTO / document JSON tags if needed |
| `internal/httpserver/channels.go` | **Modify** — `handleCreateChannel`, parse JSON body |
| `internal/httpserver/server.go` | **Modify** — `r.Post("/channels", …)` inside bearer group |
| `cmd/server/main.go` | **Modify** only if constructor wiring changes (likely none if repo already injected) |
| `internal/apiembed/openapi.yaml` | **Modify** — `POST /v1/channels`, `CreateChannelRequest`, response schema |
| `internal/httpserver/channels_test.go` | **Modify** — stub repo or fake for handler tests; optional repo test with integration tag |

## API / types

- **REST:** `POST /v1/channels` — `Authorization: Bearer`, `Content-Type: application/json`
- **Body (example):** `{"name":"general","type":"group"}`
- **201:** created channel (same shape as list items: `id`, `name`, `type`, `created_by`, `created_at`)
- **400:** validation error JSON `{"error":"..."}`
- **401:** middleware
- **500:** DB errors

## Implementation steps

1. Add repository method using **`pgxpool.Begin`** (or pool `Tx`) — insert channel with `created_by = userID`, insert `channel_members` row `(channel_id, user_id, role)`.
2. Register **`Post("/channels", …)`** next to `Get("/channels", …)` under bearer middleware.
3. Implement handler: decode JSON, validate, call repo, **`writeJSON` 201** with channel payload.
4. Update **OpenAPI** (request body, responses 201/400/401/500).
5. **Tests:** handler tests with fake/stub store; `go test ./...`, `go vet ./...`; CI race on Linux.

## Definition of done

- [x] `POST /v1/channels` creates row and membership; creator sees it in `GET /v1/channels`
- [x] OpenAPI updated
- [x] Tests green (`go test ./...`, `go vet ./...`)
- [x] **D.2** ticked in `output/backend-golang-implementation-plan.md`

## Risks / follow-ups

- **Direct** channels need two users and uniqueness rules — separate spec.
- Rate limiting (section **I**) not in this slice.

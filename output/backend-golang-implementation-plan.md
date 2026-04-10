# Internal Comm System — Golang Backend Implementation Plan

**Scope:** Backend services written in Go (chat, real-time, file metadata/presign, LiveKit token issuance), integrating with Keycloak (OIDC/JWT), PostgreSQL, Redis, MinIO, and LiveKit.  
**Audience:** Engineers tracking implementation progress.  
**Infra alignment:** Docker + Kubernetes (k3s), eventual consistency for chat, strong consistency for auth-adjacent and file metadata.

---

## 1. Backend service boundaries

| Service | Responsibility | Consistency |
|--------|----------------|-------------|
| **API Gateway (optional)** | TLS termination often at Traefik/Nginx; Go services expose HTTP/WS behind ingress | — |
| **Chat Service** | REST + WebSocket, rooms/channels, message persistence, Redis pub/sub fan-out, mentions, read receipts (later) | Eventual (real-time path); durable writes to Postgres |
| **File Service** | Presigned PUT/GET URLs, bucket policy checks, object metadata records in Postgres | Strong for metadata |
| **Admin API (optional, Go)** | If admin is not only React calling Keycloak Admin API, thin BFF for channel/role sync | Strong |

**Recommendation for 10–20 users:** Start with a **single Go binary** (modular packages: `auth`, `chat`, `file`, `livekit`) and split into two deployments later if needed. This keeps cost and ops minimal while preserving clear boundaries.

---

## 2. Suggested Go stack (2026-oriented)

| Concern | Choice | Notes |
|---------|--------|--------|
| HTTP | `chi` or `echo` or `fiber` | Pick one; `chi` is minimal and idiomatic |
| WebSocket | `nhooyr.io/websocket` or `gorilla/websocket` | Prefer context-aware, production-hardened |
| PostgreSQL | `pgx/v5` + `sqlc` or `ent` | `sqlc` for typed SQL; `ent` if you want codegen schema |
| Redis | `go-redis/v9` | Pub/Sub + optional presence TTL keys |
| JWT / OIDC | `lestrrat-go/jwx` + JWKS cache from Keycloak | Validate `iss`, `aud`, `exp`; map `sub` → internal user id |
| Config | `envconfig` or Viper | 12-factor env for K8s |
| Logging | `slog` (stdlib) or Zap | Structured JSON in prod |
| Metrics / tracing | OpenTelemetry + Prometheus exporter | Align with Prometheus/Grafana |
| Migration | `golang-migrate` or Atlas | Versioned schema for Postgres |

---

## 3. Data model (PostgreSQL)

**Principles:** Keycloak owns identity; backend stores `keycloak_sub` (unique) and optional display cache. Channels and membership for authorization on WS/REST.

- **users** — `id`, `keycloak_sub`, `display_name`, `avatar_url`, `created_at`, `updated_at`
- **channels** — `id`, `name`, `type` (direct / group), `created_by`, `created_at`
- **channel_members** — `channel_id`, `user_id`, `role`, `joined_at`
- **messages** — `id`, `channel_id`, `sender_id`, `body` (JSONB: text, mentions, attachments keys, thread parent), `created_at`, `edited_at` nullable
- **files** (metadata) — `id`, `object_key`, `bucket`, `owner_id`, `mime`, `size`, `created_at` (optional checksum)

**Indexes:** `(channel_id, created_at DESC)` for history; GIN on JSONB if search on mentions later.

---

## 4. Redis usage

- **Pub/Sub channel naming:** `channel:{channel_id}:messages` (or shard by hash if ever needed).
- **Flow:** WS client sends message → validate JWT + membership → `INSERT` message → `PUBLISH` serialized envelope → all Chat Service instances subscribed forward to local WS clients in that room.
- **Presence (Phase 5):** `SET presence:{user_id} online EX TTL` + periodic heartbeat; optional `PUBSUB` for presence events.

---

## 5. HTTP / WebSocket API (high level)

### REST (JWT Bearer)

- `GET /health`, `GET /ready` (DB + Redis checks)
- `GET /v1/channels` — list for current user
- `POST /v1/channels` — create (if permitted)
- `GET /v1/channels/{id}/messages?cursor=` — paginated history
- `POST /v1/files/presign` — upload/download presigned URL (File module)
- `POST /v1/livekit/token` — room name, identity from JWT → LiveKit access token

### WebSocket

- `GET /v1/ws?access_token=...` or `Sec-WebSocket-Protocol` bearer — establish connection
- **Client → server:** `join_channel`, `leave_channel`, `send_message`, `typing` (optional)
- **Server → client:** `message`, `message_ack`, `error`, `presence` (later)

**Protocol:** JSON messages with `type` field; version field for future compatibility.

---

## 6. Security checklist (backend)

- [ ] Validate JWT on every REST and WS upgrade (reject expired/invalid)
- [ ] Enforce channel membership before join/send/history
- [ ] Rate limit: login-adjacent endpoints (if any) + WS connect per IP/user
- [ ] Presigned URLs: short TTL, max size, content-type allowlist
- [ ] No secrets in logs; redact tokens
- [ ] CORS restricted to known admin/chat origins if browser clients exist

---

## 7. Kubernetes / Docker (Go-specific)

- Multi-stage Dockerfile: `golang:1.22+` build → `distroless` or `alpine` runtime
- Liveness/readiness probes on `/health` and `/ready`
- Config via `ConfigMap`/`Secret`: DB DSN, Redis addr, Keycloak realm URL, MinIO credentials, LiveKit API key/secret
- HPA optional: CPU-based, 2–3 replicas for HA

---

## 8. Testing strategy

- **Unit:** JWT parsing, room membership, message validation, presign URL constraints
- **Integration:** Testcontainers (Postgres + Redis + MinIO) for repository + pub/sub
- **E2E (optional):** k6 or scripted WS client for latency smoke (target &lt;100 ms same-region, not counting client render)

---

## 9. Alignment with your phased roadmap (backend-only)

| Your phase | Backend deliverables |
|------------|----------------------|
| Phase 0 | Dockerfile, Helm chart skeleton, env contract, `/health` |
| Phase 1 | JWT middleware, user upsert on first request, optional admin APIs |
| Phase 2 | WS hub, Redis pub/sub, messages API + persistence |
| Phase 3 | File presign module, `attachments` in message JSONB, metadata table |
| Phase 4 | LiveKit JWT minting endpoint, room naming convention tied to `channel_id` |
| Phase 5 | Presence, search (SQL/GIN), metrics, load test fixes, audit logs |

---

## 10. Risks (backend-focused) & mitigation

| Risk | Mitigation |
|------|------------|
| WS scale on single binary | Redis fan-out + sticky sessions or stateless instances with shared pub/sub |
| JWT clock skew | NTP on nodes; small `leeway` in validator |
| Postgres write hot spot per channel | Batch optional; for 20 users usually sufficient |
| MinIO URL leakage | Short TTL + private bucket + authz on presign |

---

# Progress checklists

Use these as the single source of truth for backend implementation status. Check items as you complete them.

## A. Foundation & repository

- [x] Initialize Go module and folder layout (`cmd/`, `internal/chat`, `internal/file`, `internal/auth`, `pkg/`)
- [x] Add structured logging (`slog`/Zap) and request ID middleware
- [x] Configuration from environment with validation (fail fast on missing secrets)
- [x] Dockerfile multi-stage + non-root user
- [x] `Makefile` or `taskfile`: `test`, `lint`, `migrate`, `run`
- [x] CI: `go test`, `golangci-lint`, image build on tag

## B. Database & migrations

- [x] Migration tool wired (migrate/Atlas)
- [x] Create tables: `users`, `channels`, `channel_members`, `messages`
- [x] Indexes for channel history pagination
- [ ] Seed script optional (dev only)

## C. Authentication (Keycloak / OIDC)

- [x] JWKS fetch + cache with refresh
- [x] JWT middleware for REST (`Authorization: Bearer`)
- [x] WS: token via query or subprotocol; same validation path
- [x] Map `sub` → `users` row (lazy create on first authenticated call)
- [x] Document required Keycloak client scopes/claims for Flutter

## D. Chat REST

- [x] `GET /v1/channels` (membership-filtered)
- [x] `POST /v1/channels` (policy: who can create)
- [x] `GET /v1/channels/{id}/messages` cursor pagination
- [ ] Authorization: only members read/write channel

## E. Real-time (WebSocket + Redis)

- [ ] Hub: register/unregister connections per `channel_id`
- [ ] Subscribe to Redis per instance; forward to local WS clients
- [ ] `send_message`: persist → publish → ACK to sender
- [ ] `join_channel` / `leave_channel` validation
- [ ] Graceful shutdown: drain WS, unsubscribe Redis

## F. File service (MinIO)

- [ ] MinIO client configuration (endpoint, creds, use SSL)
- [ ] `POST /v1/files/presign` upload (PUT) and download (GET)
- [ ] Bucket policy documented in repo (private, no public list)
- [ ] Optional `files` metadata row on successful upload callback or client confirm
- [ ] Message JSON schema includes `attachments[]` with `object_key`, `mime`, `size`

## G. LiveKit integration

- [ ] Load LiveKit API key/secret from secrets
- [ ] `POST /v1/livekit/token`: room id from `channel_id` or call session id, identity = `sub` or internal user id
- [ ] Document Flutter: obtain token before `Room.connect`
- [ ] TURN/STUN: rely on LiveKit server config (verify in staging)

## H. Observability & ops

- [x] `/health` (liveness) and `/ready` (DB + Redis + Keycloak issuer + MinIO health when configured)
- [ ] Prometheus metrics: active WS, messages/sec, Redis errors, DB latency
- [ ] OpenTelemetry traces for HTTP handlers (optional)
- [ ] Helm chart values for replicas, resources, env

## I. Hardening & performance (Phase 5)

- [ ] Rate limiting (per user/IP) on HTTP and WS connect
- [ ] Message size limits and JSON depth limits
- [ ] Load test with ~20 concurrent users; tune connection pools
- [ ] Audit log table or structured logs for admin actions (if applicable)

## J. Documentation handoff

- [x] OpenAPI/Swagger or static markdown for REST + WS message types (OpenAPI 3 at `internal/apiembed/openapi.yaml`, served at `/docs`, `/openapi.yaml`; WS payloads TBD with chat section E)
- [x] Example `.env` for local dev (no real secrets)
- [ ] Runbook: rotate LiveKit keys, DB backup restore smoke test

---

## Quick reference: milestone mapping

| Milestone | Checklist sections |
|-----------|-------------------|
| M1 — Cluster + HTTPS | A (Docker), H (health) |
| M2 — SSO working | C, B (users) |
| M3 — Real-time chat &lt;100 ms (same region) | D, E |
| M4 — Files in messages | F |
| M5 — Voice/video token path | G |
| M6 — Production-ready | H, I, J |

---

*Document version: 1.0 — Backend Golang plan derived from Internal Comm System architecture (self-hosted, Keycloak, Redis, Postgres, MinIO, LiveKit).*

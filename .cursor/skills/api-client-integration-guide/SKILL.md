---
name: api-client-integration-guide
description: Produces markdown client integration guides for web and mobile apps covering REST and WebSocket usage, OIDC/JWT, OpenAPI discovery, and environment-specific base URLs for local direct access versus TLS ingress. Use when the user asks for API integration documentation, client onboarding guides, consumer developer docs, or artifacts under output/api-guide/.
---

# API client integration guide (Internal Comm backend)

## Goal

Write a **single, copy-paste-friendly guide** so **browser and mobile** teams can integrate with this HTTP API: where to call, how to authenticate, which paths exist, and how **local dev (no ingress)** differs from **deployments behind nginx `/api` + Keycloak `/oauth`**.

## Source of truth (read first)

| Asset | Use |
|-------|-----|
| `internal/apiembed/openapi.yaml` | Canonical paths, methods, schemas, `bearerAuth`, WS token rules |
| `docs/keycloak-browser.md`, `docs/keycloak-flutter.md` | Browser vs mobile OIDC patterns, `aud`/`azp`, Web origins |
| `README.md` | Run modes, ports, ingress URLs, profile names (`app`, `ingress`) |
| `internal/httpserver/server.go` | When `/v1/*` is mounted; public routes if auth off |
| `internal/httpserver/openapi.go`, `internal/config/config.go` | `OPENAPI_PUBLIC_BASE_URL`, forwarded prefix behavior for docs/spec |
| `.env.example` | Typical env names (no real secrets) |

If OpenAPI and code disagree, **note the mismatch** in the generated guide and prefer aligning OpenAPI in the same change as routes (see `.cursor/rules/openapi-sync.mdc`).

## Output location

Write under **`output/api-guide/`** in a **dated or named subfolder**, e.g.:

- `output/api-guide/2026-04-06/` (ISO date), or  
- `output/api-guide/<brief-label>/` if the user names a release or audience.

**Required artifact:** `README.md` in that folder (main integration guide). Optional: `cheatsheet.md` for one-page URL tables only.

Do **not** commit real client secrets, refresh tokens, or sample JWTs.

## Document structure

Follow the section order and depth in [reference.md](reference.md). Minimum content:

1. **Audience** — web SPA, native mobile, or both; assumption that clients use OIDC with Keycloak as in this repo.
2. **Two deployment profiles** — side-by-side **Direct (local / published API)** vs **TLS ingress** with example bases:
   - Direct API: `http://localhost:8080` (adjust host/port if documented elsewhere).
   - Ingress: `https://<host>/api` as API root (no trailing slash on the variable); Keycloak public base `https://<host>/oauth`.
3. **Discovery** — `GET /openapi.yaml`, `GET /docs` (Swagger); explain that under ingress these are under `/api/openapi.yaml` and `/api/docs` when using the same host pattern as README.
4. **Authentication** — Bearer JWT on secured routes; `KEYCLOAK_AUDIENCE` / token `aud` or `azp`; link to Keycloak docs for obtaining tokens (authorization code / PKCE for browsers; mobile patterns per flutter doc). Password grant: mention **dev-only** if at all.
5. **REST overview** — Table or short list: health, ready, profile, channels, messages (cursor), aligned with OpenAPI **tags** and **operationId** names where helpful.
6. **WebSocket** — `GET /v1/ws` upgrade; token via query `access_token` or `Sec-WebSocket-Protocol`; give **ws://** vs **wss://** examples for direct vs ingress.
7. **CORS and TLS** — State clearly: **API CORS is not enabled by default** (see `docs/keycloak-browser.md`); browsers may need a proxy, extension, or proxy headers at the edge. Ingress uses TLS; dev certs may require trust or tooling flags.
8. **Errors and edge cases** — 401 invalid/missing token; 404 on `/v1/*` when server has no `KEYCLOAK_ISSUER` / DB; 503 on `/ready` when dependencies down.
9. **Versioning** — API is under `/v1/...` where applicable; mention OpenAPI `info.version` as contract version.

## Quality bar

- Concrete **example URLs** for both profiles (placeholders like `<access_token>` only).
- **No** invented endpoints; every documented path must exist in OpenAPI or be explicitly labeled as non-contract (e.g. Keycloak admin URLs).
- Short sentences; tables for URL comparison; link paths relative to repo (`docs/...`) where useful.
- If the user names a **production host** or **custom paths**, reflect those explicitly in the tables while keeping defaults from README.

## Optional regeneration note

Add a one-line footer in `README.md`: regenerate via this skill after OpenAPI or ingress routing changes.

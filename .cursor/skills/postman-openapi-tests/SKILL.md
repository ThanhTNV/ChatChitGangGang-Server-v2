---
name: postman-openapi-tests
description: Reads the repo OpenAPI spec and auth/docs workflows to generate Postman Collection v2.1 JSON, optional environment files, and markdown test documentation under output/postman/. Use when the user asks for Postman collections, Newman automation, API smoke tests from OpenAPI, or contract-style HTTP tests for this backend.
---

# Postman tests from OpenAPI (Internal Comm backend)

## Goal

Produce **maintainable Postman artifacts** aligned with this repo’s **canonical HTTP contract** and documented auth flows, without hard-coding secrets.

## Source of truth (read first)

| Asset | Use |
|-------|-----|
| `internal/apiembed/openapi.yaml` | Paths, methods, request/response shapes, `bearerAuth`, tags |
| `docs/keycloak-flutter.md`, `docs/keycloak-browser.md` | How to obtain `access_token`, `aud`/`azp`, WS token delivery |
| `internal/httpserver/server.go` | Confirm **when `/v1/*` is mounted** (Keycloak + DB); note 404 if auth disabled |
| `.env.example` | Suggested `HTTP_ADDR`, Keycloak URLs for environment defaults |

If OpenAPI and code disagree, **prefer fixing OpenAPI in the same change as the route** (see `.cursor/rules/openapi-sync.mdc`); for this skill, **call out mismatches** in the generated test doc.

## Output location

Write under **`output/postman/`** using a **dated or named subfolder** per run, e.g.:

- `output/postman/2026-04-10/` (ISO date) **or**
- `output/postman/<brief-label>/` if the user names a release or feature slice

Required artifacts in that folder:

| File | Purpose |
|------|---------|
| `README.md` | How to import, required env vars, **test execution order**, auth workflow, known limits (e.g. WebSocket) |
| `*.postman_collection.json` | **Collection v2.1** — folders by OpenAPI **tag** (or by path prefix if untagged) |
| `local.postman_environment.json` *(optional)* | Variables only — **no real tokens**; placeholders for `base_url`, `access_token` |

Do **not** write real client secrets or JWTs into the repo.

## Collection design rules

1. **Variables**
   - `base_url` — default `http://localhost:8080` (or from user).
   - `access_token` — empty default; user pastes token or uses OAuth2 collection auth manually.
   - Optional: `keycloak_token_url`, `keycloak_client_id` if documenting password/client-credentials flow as *placeholders* only.

2. **Requests**
   - URL: `{{base_url}}` + path from OpenAPI (e.g. `/v1/channels`).
   - **Bearer**: `Authorization: Bearer {{access_token}}` on secured operations; omit where OpenAPI has no `security`.
   - Headers: `Accept: application/json`; `Content-Type: application/json` for bodies.

3. **Test scripts (automation)**
   - On each request, add a **Tests** tab script using `pm.test` for:
     - Expected status (200, 201, 401, 400 as documented).
     - Minimal JSON shape when helpful (`pm.response.json()`, key presence).
   - Shared logic: collection-level script only when it reduces duplication (e.g. `pm.response.to.have.status` pattern).

4. **Ordering / workflows**
   - Document in `README.md` a **happy path**: e.g. `GET /health` → `GET /ready` → `GET /v1/me` (with token) → `GET /v1/channels` → `POST /v1/channels`.
   - Note **prerequisites**: migrations, `KEYCLOAK_ISSUER` set, valid `KEYCLOAK_AUDIENCE` matching token `aud`/`azp`.

5. **WebSocket**
   - Postman supports WS but differs from REST collections; either add a **separate WS request** with documented limitations **or** document **`curl` / client** using `access_token` query (see OpenAPI `/v1/ws`).

## Newman (CLI) hint

Document in `README.md`:

```bash
newman run output/postman/<folder>/internal-comm.postman_collection.json \
  -e output/postman/<folder>/local.postman_environment.json \
  --env-var "access_token=$ACCESS_TOKEN"
```

User supplies `ACCESS_TOKEN` from their IdP.

## Quality bar

- Collection must **validate as JSON** and import in Postman without manual path fixes.
- Every secured endpoint in OpenAPI should appear unless explicitly excluded with a reason in `README.md`.
- Keep collection **IDs stable within a file** (`uuid`-style strings) so git diffs are predictable when regenerating.

## Additional format details

See [reference.md](reference.md) for Collection v2.1 skeleton and naming conventions.

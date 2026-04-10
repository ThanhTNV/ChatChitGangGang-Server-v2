# Internal Comm Backend — API integration guide

**Contract:** OpenAPI 3 **`info.version` 0.1.0** — [`internal/apiembed/openapi.yaml`](../../../internal/apiembed/openapi.yaml).  
**Audience:** Web (SPA / BFF) and native mobile clients using **Keycloak OIDC** JWTs against this Go API.

**Compose layout:** [`docker-compose.yml`](../../../docker-compose.yml) = direct HTTP (**no** TLS nginx). Add [`docker-compose.ingress.yml`](../../../docker-compose.ingress.yml) for **`https://localhost/api`** + **`https://localhost/oauth`**. Command:

`docker compose -f docker-compose.yml -f docker-compose.ingress.yml --profile app up -d --build`

---

## 1. API base URLs (pick one column)

Paths below are **relative to the API base** (not Keycloak).

| Concern | Direct (Compose or `go run`, no ingress) | TLS ingress (merged Compose + nginx) |
|--------|------------------------------------------|--------------------------------------|
| **API base** | `http://localhost:8080` | `https://localhost/api` — **no trailing slash** on stored config |
| **Health** | `GET http://localhost:8080/health` | `GET https://localhost/api/health` |
| **Ready** | `GET http://localhost:8080/ready` | `GET https://localhost/api/ready` |
| **OpenAPI** | `GET http://localhost:8080/openapi.yaml` | `GET https://localhost/api/openapi.yaml` |
| **Swagger** | `GET http://localhost:8080/docs` | `GET https://localhost/api/docs` |
| **Secured REST** | `GET http://localhost:8080/v1/me` | `GET https://localhost/api/v1/me` |
| **WebSocket** | `ws://localhost:8080/v1/ws?access_token=<JWT>` | `wss://localhost/api/v1/ws?access_token=<JWT>` |

The Go process always sees **`/health`**, **`/v1/...`**, etc. Ingress strips the **`/api`** prefix before proxying.

**OpenAPI “servers” / Swagger “Try it out”:** The server may set **`OPENAPI_PUBLIC_BASE_URL`** (ingress overlay sets **`https://localhost/api`**) or derive from **`Host`** + **`X-Forwarded-*`** (see [`internal/httpserver/openapi.go`](../../../internal/httpserver/openapi.go), [`.env.example`](../../../.env.example)).

---

## 2. Keycloak — direct vs ingress (must match `KEYCLOAK_ISSUER`)

JWT claim **`iss`** must equal the API’s **`KEYCLOAK_ISSUER`**. Use the discovery document whose **`issuer`** field you will align with the API.

| Mode | Keycloak URLs (realm `master` examples) | Typical `iss` |
|------|-------------------------------------------|---------------|
| **Direct** — `docker-compose.yml` only | Discovery: `http://localhost:8090/realms/master/.well-known/openid-configuration` · Admin: `http://localhost:8090/admin/` | `http://localhost:8090/realms/master` |
| **Ingress** — merge `docker-compose.ingress.yml` | Discovery: `https://localhost/oauth/realms/master/.well-known/openid-configuration` · Admin: `https://localhost/oauth/admin/...` | `https://localhost/oauth/realms/master` |
| **Ingress stack, hitting published :8090** | `http://localhost:8090/oauth/realms/...` (Keycloak uses **`KC_HTTP_RELATIVE_PATH=/oauth`**) | Still align tokens with **`https://localhost/oauth/realms/master`** if users log in via HTTPS |

Details and CORS: [`docs/keycloak-browser.md`](../../../docs/keycloak-browser.md), mobile: [`docs/keycloak-flutter.md`](../../../docs/keycloak-flutter.md).

**Redirect:** Opening **`https://localhost/oauth`** may redirect to **`/oauth/admin/...`** — expected Keycloak behavior.

---

## 3. Authentication

- **REST (secured routes):** `Authorization: Bearer <access_token>` · `Accept: application/json` · JSON bodies: `Content-Type: application/json`.
- **Obtain tokens:** Browser — authorization code + **PKCE**; mobile — same pattern per flutter doc. **Resource-owner password grant** is **dev-only** (e.g. Postman), not for production SPAs.
- **Audience:** API **`KEYCLOAK_AUDIENCE`** must match JWT **`aud`** and/or **`azp`** (often the OIDC **client id**). See flutter doc.

---

## 4. Discovery and contract

| Resource | Path on API | Notes |
|----------|-------------|--------|
| OpenAPI YAML | **`GET /openapi.yaml`** | Machine-readable contract |
| Swagger UI | **`GET /docs`** | Loads spec from `openapi.yaml` relative to page URL; **not** listed under `paths` in the embedded YAML — still served by the app |

---

## 5. REST summary (OpenAPI)

All **`/v1/...`** require a valid Bearer token when the route is mounted.

| Tag | Method | Path | Notes |
|-----|--------|------|--------|
| Health | GET | `/health` | `{"status":"ok"}` |
| Health | GET | `/ready` | **200** or **503**; `checks` per dependency |
| Health | GET | `/openapi.yaml` | Raw spec |
| Profile | GET | `/v1/me` | **401** if bad/missing token |
| Chat | GET | `/v1/channels` | Member’s channels |
| Chat | POST | `/v1/channels` | Create **group** channel — **201** |
| Chat | GET | `/v1/channels/{channelID}/messages` | Query **`limit`** (1–100, default 50), optional **`cursor`** from **`next_cursor`** |

---

## 6. WebSocket

- **`GET /v1/ws`** (upgrade). Token: query **`access_token=<JWT>`** and/or **`Sec-WebSocket-Protocol: bearer.<JWT>`** (single value; do not split JWT on commas).
- **401** before upgrade: **plain text** body.
- Use **`wss://`** with HTTPS / ingress.

---

## 7. CORS and TLS

- **Go API:** no default **CORS** middleware — browser SPAs on another origin often need a **dev proxy**, **BFF**, or edge CORS. See [`docs/keycloak-browser.md`](../../../docs/keycloak-browser.md).
- **Keycloak:** configure **Web origins** on the client for calls **to Keycloak** from the browser.
- **Ingress:** dev **self-signed** certs — trust the CA or disable verification in tooling (not in production).

---

## 8. When `/v1/*` returns 404

**`/v1`** is mounted only when the server has auth + DB wiring (**`KEYCLOAK_ISSUER`** set and valid database path for user sync per [`internal/httpserver/server.go`](../../../internal/httpserver/server.go)). If issuer is unset, only health, ready, and docs routes exist.

---

## 9. Versioning

HTTP API surface under **`/v1/...`**. Contract version is OpenAPI **`info.version`** (currently **0.1.0**).

---

## 10. Server env (reference for integrators)

See [`.env.example`](../../../.env.example). Common auth-related vars: **`KEYCLOAK_ISSUER`**, **`KEYCLOAK_AUDIENCE`**, **`KEYCLOAK_JWKS_URL`**, **`KEYCLOAK_READY_URL`**, **`OPENAPI_PUBLIC_BASE_URL`**.

---

_Regenerate with the **api-client-integration-guide** Cursor skill after OpenAPI or ingress routing changes._

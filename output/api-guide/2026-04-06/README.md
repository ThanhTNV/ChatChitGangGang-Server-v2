# Internal Comm Backend — API integration guide

**Contract version:** OpenAPI `info.version` **0.1.0** (`internal/apiembed/openapi.yaml`).  
**Audience:** Web (SPA or BFF) and native mobile clients that obtain **Keycloak OIDC** access tokens and call this Go API.

HTTP routes under **`/v1/...`** exist only when the server is configured with **`KEYCLOAK_ISSUER`** and related auth + database settings (see below). Otherwise those paths return **404**.

---

## 1. Two ways to reach the API

Use **one** column consistently: paths below are **relative to the API base** (not including Keycloak).

| Concern | Direct API (local / published port) | TLS ingress (nginx, repo Compose) |
|--------|--------------------------------------|-----------------------------------|
| **API base** | `http://localhost:8080` | `https://<host>/api` — **no trailing slash** on the stored base URL |
| **Liveness** | `GET http://localhost:8080/health` | `GET https://<host>/api/health` |
| **Readiness** | `GET http://localhost:8080/ready` | `GET https://<host>/api/ready` |
| **OpenAPI YAML** | `GET http://localhost:8080/openapi.yaml` | `GET https://<host>/api/openapi.yaml` |
| **Swagger UI** | `GET http://localhost:8080/docs` | `GET https://<host>/api/docs` |
| **Example REST** | `GET http://localhost:8080/v1/me` | `GET https://<host>/api/v1/me` |
| **WebSocket** | `ws://localhost:8080/v1/ws?access_token=<JWT>` | `wss://<host>/api/v1/ws?access_token=<JWT>` |

**Ingress details:** Merge **`docker-compose.ingress.yml`**: **`/api/...`** is stripped to the Go API; **`/oauth/...`** is forwarded to Keycloak with the **`/oauth`** prefix kept (**`KC_HTTP_RELATIVE_PATH=/oauth`**). See root **`README.md`**.

**Host `go run`:** The API listens on **`HTTP_ADDR`** (often `:8080`). That process is **not** behind `/api` unless you add your own reverse proxy. Do not prepend `/api` when calling `http://localhost:8080` directly.

**OpenAPI “servers” in Swagger:** The server may rewrite the embedded **`servers`** entry using **`OPENAPI_PUBLIC_BASE_URL`**, or derive it from the request (**`X-Forwarded-Proto`**, **`X-Forwarded-Host`**, **`X-Forwarded-Prefix`**, e.g. `/api`). That keeps “Try it out” URLs aligned with how you browse **`/docs`**.

---

## 2. Keycloak (OIDC) — issuer and tokens

### Direct vs ingress (same Keycloak process)

| Mode | How users reach Keycloak | Example issuer / discovery |
|------|--------------------------|----------------------------|
| **No ingress** | `http://localhost:8090` | `http://localhost:8090/realms/master` — discovery at **`/realms/master/.well-known/openid-configuration`** |
| **TLS ingress** | `https://localhost/oauth/...` | `https://localhost/oauth/realms/master` — discovery under **`/oauth/realms/...`** |

**Important:** JWT **`iss`** must match **`KEYCLOAK_ISSUER`** on the API. Pick **one** client entry point per environment (direct **or** ingress) and set the API issuer accordingly. Switching modes requires matching **`KEYCLOAK_ISSUER`** (and usually re-authenticating so tokens get the new **`iss`**).

More detail: [`docs/keycloak-browser.md`](../../../docs/keycloak-browser.md).

### Audience

The API validates **`KEYCLOAK_AUDIENCE`** against the JWT **`aud`** and/or **`azp`** (authorized party / client id). Public clients often need **`KEYCLOAK_AUDIENCE`** set to the **client id**. Details: [`docs/keycloak-flutter.md`](../../../docs/keycloak-flutter.md), [`docs/keycloak-browser.md`](../../../docs/keycloak-browser.md).

### Obtaining tokens (client apps)

| Client | Recommended approach |
|--------|----------------------|
| **Browser / SPA** | Authorization code **with PKCE**; optional refresh token rotation. See [`docs/keycloak-browser.md`](../../../docs/keycloak-browser.md). |
| **Mobile (e.g. Flutter)** | Authorization code with PKCE; see [`docs/keycloak-flutter.md`](../../../docs/keycloak-flutter.md). |
| **Dev / Postman only** | Resource-owner password grant — **not** for production mobile/SPA builds; requires **Direct access grants** on the client. |

Never ship **client secrets** in public SPA or mobile binaries.

### Calling secured REST endpoints

```http
Authorization: Bearer <access_token>
Accept: application/json
```

JSON bodies: **`Content-Type: application/json`**.

---

## 3. Discovery and documentation

| Resource | Path (on API) | Notes |
|----------|---------------|--------|
| Machine-readable contract | **`GET /openapi.yaml`** | OpenAPI 3 YAML |
| Human exploration | **`GET /docs`** | Swagger UI loads spec from **`openapi.yaml`** relative to the page URL |

**OpenAPI vs runtime:** The YAML file lists **`/health`**, **`/ready`**, **`/openapi.yaml`**, **`/v1/me`**, **`/v1/channels`**, **`/v1/channels/{channelID}/messages`**, **`/v1/ws`**. **`/docs`** is served by the server but is **not** a separate `paths` entry in the embedded YAML; treat it as a first-class doc URL for integrators.

---

## 4. REST endpoints (summary)

All **`/v1/...`** operations below require a valid Bearer token when the route is mounted.

| Tag | Method | Path | Purpose |
|-----|--------|------|---------|
| Health | GET | `/health` | Liveness — JSON `{"status":"ok"}` |
| Health | GET | `/ready` | Readiness — **200** or **503** with per-dependency `checks` |
| Health | GET | `/openapi.yaml` | Raw OpenAPI document |
| Profile | GET | `/v1/me` | Current user (upsert from JWT); **401** if invalid/missing token |
| Chat | GET | `/v1/channels` | List channels for the member |
| Chat | POST | `/v1/channels` | Create **group** channel; **201** — body `CreateChannelRequest` (`name`, optional `type`: `group`) |
| Chat | GET | `/v1/channels/{channelID}/messages` | Cursor-paginated messages; query `limit` (1–100, default 50), optional `cursor` |

**Pagination (messages):** First request omit **`cursor`**. If the response includes **`next_cursor`**, pass it as **`cursor`** on the next **`GET`** to fetch older messages. Newest-first ordering per OpenAPI.

**Errors (typical):** JSON **`{"error":"..."}`** for many failures; see each operation in OpenAPI for **400** / **401** / **404** / **500**.

---

## 5. WebSocket

- **Path:** **`GET /v1/ws`** (WebSocket upgrade).
- **Auth:** Browsers and mobile stacks often cannot set **`Authorization`** on the WS handshake the same way as REST. This API accepts:
  - Query: **`access_token=<JWT>`**, and/or
  - Header **`Sec-WebSocket-Protocol`** as a **single** value **`bearer.<JWT>`** (do not split the JWT on commas).
- **Failure before upgrade:** **401** with **plain text** body (not JSON).
- **URL scheme:** Use **`ws://`** for direct HTTP API bases and **`wss://`** when the page/API is served over HTTPS (including ingress).

---

## 6. CORS, TLS, and security notes

- **CORS:** The Go server does **not** enable CORS middleware by default. A SPA on another origin (e.g. `http://localhost:3000` → `http://localhost:8080`) may need a **dev proxy**, **reverse-proxy CORS**, or a **BFF**. See [`docs/keycloak-browser.md`](../../../docs/keycloak-browser.md) (*Cross-origin notes*).
- **Keycloak CORS:** Configure **Web origins** on the Keycloak client for calls **to Keycloak** from the browser — separate from API CORS.
- **TLS:** Ingress uses HTTPS; **self-signed** dev certificates require trusting the CA or disabling verification in tooling (not in production).
- **Forwarded headers:** OpenAPI base URL derivation trusts **`X-Forwarded-*`** when present. In production, ensure only your edge sets these headers.

---

## 7. When `/v1/*` is missing (404)

The server mounts **`/v1`** only when auth wiring is complete (Bearer middleware, validator, user sync, and stores). In code terms this requires **`KEYCLOAK_ISSUER`** and a working **`DATABASE_URL`** path for user upsert. If **`KEYCLOAK_ISSUER`** is unset, **`/health`**, **`/ready`**, **`/openapi.yaml`**, and **`/docs`** still work; **`/v1/*`** does not.

---

## 8. Troubleshooting

| Symptom | Likely cause |
|---------|----------------|
| **401** on `/v1/...` | Expired token, wrong signing keys, or **`aud`/`azp`** mismatch with **`KEYCLOAK_AUDIENCE`** |
| **404** on all `/v1/...` | Auth stack not enabled on server |
| **503** on `/ready` | Postgres, Redis, Keycloak, or MinIO check failed (see response `checks`) |
| **Swagger “Try it out” wrong host** | Set **`OPENAPI_PUBLIC_BASE_URL`** or fix proxy forwarded headers |
| **Redirect loop on Keycloak admin (ingress)** | Public hostname / HTTPS proxy settings — see root **`README.md`** |

---

## 9. Environment variables (API server)

Integrators usually do not set these; they help you understand deployment requirements. Full list: **[`.env.example`](../../../.env.example)**. Relevant to auth:

- **`KEYCLOAK_ISSUER`**, **`KEYCLOAK_AUDIENCE`**, **`KEYCLOAK_JWKS_URL`** (optional), **`KEYCLOAK_READY_URL`** (optional)
- **`DATABASE_URL`**
- **`OPENAPI_PUBLIC_BASE_URL`** (optional)

---

_Regenerate this guide with the **api-client-integration-guide** Cursor skill after OpenAPI or ingress routing changes._

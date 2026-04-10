# Postman — Internal Comm Backend (generated from OpenAPI)

**Source:** `internal/apiembed/openapi.yaml` (v0.1.0)  
**Artifacts:** `internal-comm.postman_collection.json`, `local.postman_environment.json`, `ingress.postman_environment.json`

## Import

1. Postman → **Import** → `internal-comm.postman_collection.json`.
2. Import **one** environment (see below) and select it in the workspace dropdown.
3. Set **Keycloak** username/password locally (not committed in git).

## Choosing an environment

| File | When to use | `base_url` | `keycloak_token_url` |
|------|-------------|------------|------------------------|
| **`local.postman_environment.json`** | API on the host (`make run`) or **published :8080** (Compose `app` profile) | `http://localhost:8080` | `http://localhost:8090/realms/master/.../token` (direct Keycloak; no `/oauth` on :8090) |
| **`ingress.postman_environment.json`** | **nginx** TLS ingress (`--profile ingress`): API under **`/api`**, Keycloak under **`/oauth`** | `https://localhost/api` (no trailing slash) | `https://localhost/oauth/realms/master/.../token` |

**TLS / self-signed:** For `https://localhost`, turn off **SSL certificate verification** in Postman (Settings → General) or add your CA, or Newman `--insecure`.

**Alignment with the server:** The API uses **`OPENAPI_PUBLIC_BASE_URL`** or **`X-Forwarded-Prefix: /api`** so OpenAPI/Swagger match the same base you use in Postman.

## Environment variables

| Variable | Purpose |
|----------|---------|
| `base_url` | Root for all REST calls in the collection (`/health`, `/v1/...`). Must match how you reach the Go process (see table above). |
| `access_token` | JWT; filled by **Keycloak auth → Get token** or paste / Newman. |
| `channel_id` | UUID for message routes; set on **Create group channel** **201**. |
| `messages_cursor` | From **`next_cursor`** on list messages; used by **List channel messages (cursor)**. |
| `keycloak_token_url` | OIDC token endpoint: **direct** `http://localhost:8090/realms/.../token`; **ingress** `https://localhost/oauth/realms/.../token`. |
| `keycloak_client_id` | Client id; **Direct access grants** on for password grant. |
| `keycloak_username` / `keycloak_password` | Set locally only. |

## Prerequisites

- API reachable at the **`base_url`** you selected.
- **Postgres** + migrations when using `/ready` with `DATABASE_URL`.
- **`/v1/*`** only if **`KEYCLOAK_ISSUER`** + **`DATABASE_URL`** are set on the server; otherwise **404** on secured routes.
- JWT **audience** matches **`KEYCLOAK_AUDIENCE`**.

## Suggested execution order (happy path)

1. **Health → Get health** — **200** `{"status":"ok"}`.
2. **Health → Get ready** — **200** or **503**.
3. **Health → Get OpenAPI spec** — **200** YAML.
4. **Health → Get docs (Swagger UI)** — **200** HTML (optional smoke).
5. Set **Keycloak** credentials → **Keycloak auth → Get token (password grant)** — **200**; saves **`access_token`**.
6. **Profile → Get me** — **200** with `user_id`, `sub`.
7. **Chat → List channels** — **200** with `channels`.
8. **Chat → Create group channel** — **201**; saves **`channel_id`**.
9. **Chat → List channel messages** — **200**; may set **`messages_cursor`**.
10. **Chat → List channel messages (cursor)** — optional; needs **`messages_cursor`** or expect **400**.
11. **WebSocket → WS entry (no token)** — **401** or **404**.

**WebSocket (real 101):** Postman **WebSocket** request:

- Direct: `ws://localhost:8080/v1/ws?access_token=<JWT>`
- Ingress: `wss://localhost/api/v1/ws?access_token=<JWT>`

## Newman (CLI)

**Direct API**

```bash
newman run output/postman/2026-04-10/internal-comm.postman_collection.json \
  -e output/postman/2026-04-10/local.postman_environment.json \
  --env-var "access_token=$ACCESS_TOKEN" \
  --env-var "keycloak_username=$KEYCLOAK_USER" \
  --env-var "keycloak_password=$KEYCLOAK_PASS"
```

**Ingress (self-signed TLS)**

```bash
newman run output/postman/2026-04-10/internal-comm.postman_collection.json \
  -e output/postman/2026-04-10/ingress.postman_environment.json \
  --insecure \
  --env-var "access_token=$ACCESS_TOKEN" \
  --env-var "keycloak_username=$KEYCLOAK_USER" \
  --env-var "keycloak_password=$KEYCLOAK_PASS"
```

Optional: `--env-var "channel_id=<uuid>"`.

## Regenerating

Re-run the **postman-openapi-tests** skill after OpenAPI or routing changes.

## OpenAPI vs runtime

- All secured OpenAPI paths are in the collection, including **`GET /v1/channels/{channelID}/messages`**.
- **`/docs`** is included as **Get docs (Swagger UI)** for smoke tests; interactive use in a browser is fine too.

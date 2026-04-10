# Template: `output/api-guide/<folder>/README.md`

Use this outline. Adjust headings for clarity; keep **Direct** vs **Ingress** visible early.

```markdown
# [Product name] — API integration guide

**Contract:** OpenAPI 3 (`internal/apiembed/openapi.yaml` in repo).  
**Audience:** [web / mobile / both]

## 1. Environments at a glance

| Concern | Local / direct API | TLS ingress (nginx) |
|--------|--------------------|---------------------|
| API base | `http://localhost:8080` | `https://<host>/api` (no trailing slash) |
| OpenAPI spec | `GET {api}/openapi.yaml` | `GET https://<host>/api/openapi.yaml` |
| Swagger UI | `GET {api}/docs` | `GET https://<host>/api/docs` |
| Keycloak realm (example) | `http://localhost:8090/realms/<realm>` | `https://<host>/oauth/realms/<realm>` |
| WebSocket | `ws://localhost:8080/v1/ws?...` | `wss://<host>/api/v1/ws?...` |

Notes: Host-run `go run` uses **direct** port **8080**; ingress proxies **container** API under **`/api`**. Align token **`iss`** with server **`KEYCLOAK_ISSUER`** (see repo README).

## 2. Prerequisites

- Server has **`KEYCLOAK_ISSUER`** + **`DATABASE_URL`** for **`/v1/*`**; otherwise those routes return **404**.
- JWT **audience** matches server **`KEYCLOAK_AUDIENCE`** (often client id as `azp`).

## 3. Authentication (OIDC / JWT)

- Secured REST: header `Authorization: Bearer <access_token>`.
- Obtain tokens via your IdP flow (browser: PKCE; mobile: app-specific). See `docs/keycloak-browser.md` and `docs/keycloak-flutter.md`.
- Do not embed client secrets in mobile or public SPA builds.

## 4. REST endpoints (summary)

[List from OpenAPI: Health, Ready, Profile, Chat, … — method, path, auth required, one-line purpose.]

## 5. Pagination (messages)

- First page: `GET /v1/channels/{channelID}/messages?limit=...`
- Next page: same path with `cursor=<next_cursor>` from previous response.

## 6. WebSocket

- Connect with validated JWT: query `access_token=<JWT>` and/or protocol per OpenAPI.
- Use **wss** when the page is **https** or behind ingress.

## 7. Browser-specific: CORS

- This API does not add CORS headers by default. Options: same-origin proxy, reverse proxy CORS, or server change — see `docs/keycloak-browser.md`.

## 8. Operational checks

- `GET /health` — liveness.
- `GET /ready` — dependencies; may be **503** during startup.

## 9. Troubleshooting

- **401** — missing/expired/wrong-audience token.
- **404** on `/v1/...` — auth stack not enabled on server.
- **Redirect / TLS issues** with Keycloak behind ingress — see repo README (Keycloak hostname / proxy).

---

_Generated for client teams; regenerate when OpenAPI or ingress docs change._
```

## Naming

- Primary file: **`README.md`** in the dated subfolder.
- Optional **`cheatsheet.md`**: only URL tables + auth header line.

# Keycloak + Flutter (OIDC) for Internal Comm API

This backend validates **JWT access tokens** from Keycloak using **JWKS** (`KEYCLOAK_JWKS_URL` or derived from `KEYCLOAK_ISSUER`).

## Environment on the API

| Variable | Purpose |
|----------|---------|
| `KEYCLOAK_ISSUER` | Must match JWT `iss` exactly. **Direct Keycloak:** `http://localhost:8090/realms/<realm>`. **TLS ingress:** `https://localhost/oauth/realms/<realm>` (see repo `README.md`). If **unset**, `/v1/*` routes are **not** mounted (auth disabled). |
| `KEYCLOAK_AUDIENCE` | Expected **resource** audience **or** OAuth2 **`azp`** (authorized party / client id). Public mobile clients often only have `azp` set to the client id — use that value here. |
| `KEYCLOAK_JWKS_URL` | Optional. Default: `{issuer}/protocol/openid-connect/certs` |
| `DATABASE_URL` | **Required** when `KEYCLOAK_ISSUER` is set (user upsert on each authenticated call). |

## Keycloak client (Flutter / public)

1. Create a **realm** and an **OpenID Connect** client (e.g. `internal-comm-mobile`).
2. **Access type:** public (or confidential if you use a secret — not typical for mobile).
3. **Standard flow:** enable Authorization Code with PKCE (recommended for Flutter).
4. **Valid redirect URIs:** your app scheme / loopback for dev.
5. **Web origins:** if you use a web build, add origins as needed.

## Scopes

Request at least:

- `openid`
- `profile` (for `preferred_username`, `name`)
- `email`

The API reads:

- `sub` (Keycloak subject) → `users.keycloak_sub`
- `preferred_username` or `email` → `users.display_name` (best-effort)

## Audience vs `azp`

- **Confidential** clients / some tokens include `aud` containing your **resource server** audience (e.g. `internal-comm-api`). Set `KEYCLOAK_AUDIENCE` to that value.
- **Public** Flutter clients often receive access tokens where **`aud`** is `account` or similar, while **`azp`** is the **client id**. This API treats a token as valid if **either** `aud` contains `KEYCLOAK_AUDIENCE` **or** `azp` equals `KEYCLOAK_AUDIENCE`.

Create a **client scope** or **mapper** if you need a dedicated `aud` value for the API; otherwise use the **client id** as `KEYCLOAK_AUDIENCE`.

## Calling the API

### REST

```http
GET /v1/me
Authorization: Bearer <access_token>
```

### WebSocket

Preferred: query parameter (avoids commas inside JWT in headers):

`GET /v1/ws?access_token=<access_token>`

Alternative: `Sec-WebSocket-Protocol` value `bearer.<access_token>` (single protocol string; do not split the JWT on commas).

## Keycloak admin (local compose)

- Admin console (direct): `http://localhost:8090/admin/` (ingress: `https://localhost/oauth/admin/`). Default bootstrap user from `docker-compose.yml`.

After creating the realm, set `KEYCLOAK_ISSUER` to match how clients obtain tokens: `http://localhost:8090/realms/<your-realm>` (direct) or `https://<host>/oauth/realms/<your-realm>` (ingress).

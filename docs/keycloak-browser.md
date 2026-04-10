# Keycloak + browser (OIDC session & tokens)

This complements [keycloak-flutter.md](./keycloak-flutter.md). The Internal Comm API still validates **JWT access tokens** the same way; the difference is how a **browser** obtains tokens, keeps a session alive, and signs out.

## Concepts: SSO session vs API tokens

| Layer | What it is | Where it lives |
|--------|------------|----------------|
| **Keycloak SSO session** | User is logged in to Keycloak (cookie on the **Keycloak host**, e.g. `localhost:8090`) | Browser cookies on IdP origin |
| **Access token** | Short-lived JWT your **API** accepts (`Authorization: Bearer …`) | Your app (see storage below) |
| **Refresh token** | Lets the app get new access tokens without another login (if enabled on the client) | Your app (prefer secure storage patterns) |

A user can have a **valid SSO session** at Keycloak but **no access token** in your SPA until you run the code flow again or use silent refresh (if you implement it).

## OIDC discovery URL (this repo — direct vs ingress)

- **`docker-compose.yml` only:** Keycloak uses **root** paths on **8090** — **`/realms/...`**, **`/admin/...`**.
- **`docker-compose.yml` + `docker-compose.ingress.yml`:** Keycloak runs with **`KC_HTTP_RELATIVE_PATH=/oauth`**. Nginx forwards **`https://localhost/oauth/...`** to **`http://keycloak:8080/oauth/...`** (prefix **kept**). Published **8090** then serves **`http://localhost:8090/oauth/realms/...`** (not `/realms/...` at the root).

| How clients reach Keycloak | Discovery (realm `master`) | Typical JWT `iss` (must match API `KEYCLOAK_ISSUER`) |
|----------------------------|------------------------------|------------------------------------------------------|
| **Direct** (`docker-compose.yml` only) | `http://localhost:8090/realms/master/.well-known/openid-configuration` | `http://localhost:8090/realms/master` |
| **TLS ingress** (merged compose) | `https://localhost/oauth/realms/master/.well-known/openid-configuration` | `https://localhost/oauth/realms/master` |
| **Same host, published 8090 + ingress stack** | `http://localhost:8090/oauth/realms/master/.well-known/openid-configuration` | still **`https://localhost/oauth/realms/master`** if users authenticate via HTTPS |

Use the discovery URL that matches how **your SPA** talks to Keycloak, and set the API **`KEYCLOAK_ISSUER`** to the same **`issuer`** value from that JSON. You cannot mix token sources (direct vs ingress) against a single API issuer without re-login or a multi-issuer API change.

## CORS from your SPA origin

Browser calls from **`http://localhost:8081`** (or any dev port) to Keycloak are cross-origin. Keycloak only adds **`Access-Control-Allow-Origin`** when the request **Origin** is allowed on the client:

1. Keycloak Admin → **Clients** → your SPA client → **Settings**.
2. **Web origins** — add **`http://localhost:8081`** (no path; exact origin). Use `+` to add multiple dev origins.
3. Save. Retry discovery or token calls from the SPA.

If the discovery URL is wrong (404), the response may not include CORS headers, and the console will show a **CORS** error even though the root cause is **404**.

## Keycloak client (browser / SPA)

1. Create an **OpenID Connect** client (e.g. `internal-comm-web`).
2. **Client type:** public (typical for SPAs) or confidential if you use a **backend-for-frontend (BFF)** that holds secrets.
3. **Standard flow:** enable **Standard flow** (authorization code).
4. **PKCE:** required for public clients; enable **Proof Key for Code Exchange**.
5. **Valid redirect URIs:** exact URLs for your app (e.g. `http://localhost:3000/*`, production `https://app.example.com/*`).
6. **Web origins:** add your SPA origin(s) for CORS when using the Keycloak JS adapter against Keycloak endpoints from the browser.
7. **Refresh tokens:** if you need long-lived sessions without redirecting the user, enable **Client → OpenID Connect Compatibility Modes** options as needed (e.g. use refresh tokens for SPAs) and prefer **refresh token rotation** in production.

Align **`KEYCLOAK_AUDIENCE`** on the API with your token’s **`aud`** or **`azp`** — same rules as in [keycloak-flutter.md](./keycloak-flutter.md).

## Recommended flows

### SPA (static site + API on another host)

1. Redirect user to Keycloak **authorize** endpoint with **response_type=code**, **code_challenge** / **code_challenge_method=S256** (PKCE).
2. User signs in; Keycloak redirects back with **`code`**.
3. Exchange **`code`** at the **token** endpoint for **access_token** (and optionally **refresh_token**).
4. Call the API with:

   ```http
   Authorization: Bearer <access_token>
   ```

5. Before **access_token** expires, use **refresh_token** (if issued) to obtain a new pair, or send the user through a **silent** authorize request if you rely on SSO cookie only (iframe / redirect nuances apply; many teams prefer refresh tokens with rotation).

**Libraries:** e.g. [oidc-client-ts](https://github.com/authts/oidc-client-ts) or Keycloak’s own JS adapter — both wrap authorize + PKCE + refresh patterns.

### BFF (cookie session in browser, tokens on server)

The browser never stores access or refresh tokens. Your **same-origin** backend exchanges the code, stores tokens server-side (session), and attaches `Authorization: Bearer` when calling the Go API. Logout clears the server session and optionally redirects to Keycloak **end session**.

This avoids exposing refresh tokens to JavaScript and simplifies some CSRF/XSS tradeoffs at the cost of more backend code.

## Session management (practical checklist)

- **Access token expiry:** decode `exp` (or use library helpers); refresh or re-authenticate before calling the API with an expired token.
- **Refresh failure** (revoked user, expired refresh): treat as logged out — clear app state and send user to login.
- **Silent renew:** if using SSO cookie + hidden iframe or `prompt=none`, be aware of third-party cookie restrictions; test in your target browsers.
- **Tab sync:** if you store session in `localStorage`, listen for `storage` events or `BroadcastChannel` so logout in one tab clears others.
- **Focus / visibility:** on `document.visibilitychange`, you may optionally re-check expiry or ping userinfo — keep it rate-limited.

## Logout

1. **App state:** clear in-memory tokens and any client-side storage you use for tokens or user profile cache.
2. **Keycloak SSO:** redirect the browser to Keycloak’s **end session** URL (from OpenID Configuration `end_session_endpoint`), with **`post_logout_redirect_uri`** (must be allowed on the client) and typically **`id_token_hint`** if you have the ID token — so Keycloak clears the SSO session and returns the user to your app.

Without step 2, the user may still be SSO-logged in at Keycloak and get a new code without credentials on the next login redirect.

## Calling this backend from the browser

Same as mobile:

- **REST:** `Authorization: Bearer <access_token>` (e.g. `GET /v1/me`, `GET /v1/channels`).
- **WebSocket:** prefer `GET /v1/ws?access_token=<access_token>` or `Sec-WebSocket-Protocol: bearer.<access_token>`.

API environment variables are documented in [keycloak-flutter.md](./keycloak-flutter.md) and `.env.example`.

## Cross-origin notes

If the SPA is served from a **different origin** than the Go API (e.g. `localhost:3000` → `localhost:8080`), the browser enforces **CORS**. This repository’s server does not add CORS middleware by default; for local dev you can:

- Proxy API requests through the SPA dev server, or  
- Terminate TLS at a reverse proxy that sets CORS headers, or  
- Add CORS in the Go server in a follow-up change.

Keycloak admin console: **direct** `http://localhost:8090/admin/` — **ingress** `https://localhost/oauth/admin/` — see [keycloak-flutter.md](./keycloak-flutter.md).

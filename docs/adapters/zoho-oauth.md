# Zoho OAuth Adapter Specification

[日本語版](zoho-oauth.ja.md)

Shared rules: [Scenario Policy](../scenario-policy.md).

This document describes the Mockport `zoho-oauth` adapter contract. It is not a copy of Zoho's documentation and does not claim full Zoho OAuth compatibility. It provides a minimal, deterministic local emulation of the Zoho OAuth2 authorization-code flow so an application can complete login locally without reaching real Zoho.

## Scope

The `zoho-oauth` adapter mocks only the three endpoints a Zoho OAuth client actually calls:

- Authorization redirect (no login screen; immediate redirect).
- Authorization-code token exchange.
- Authenticated user info lookup using the Zoho-specific auth scheme.

It does not implement the rest of the Zoho API surface.

## Base Path

Default base path (the value an application sets as `ZOHO_AUTH_BASE_URL`):

```text
/zoho
```

Point the application's `ZOHO_AUTH_BASE_URL` at this Mockport base path to complete login locally.

Example config:

```yaml
adapters:
  zoho-oauth:
    enabled: true
    base_path: /zoho
    scenario: oauth_success
    fake_secret: mockport_zoho_secret
```

## Official Reference Map

These links are references for behavior shape only; Mockport remains a deterministic local emulator.

| Mockport surface | Official reference |
| --- | --- |
| OAuth authorize redirect and token exchange | `https://www.zoho.com/accounts/protocol/oauth/web-apps/authorization.html` |
| User info with the `Zoho-oauthtoken` scheme | `https://www.zoho.com/accounts/protocol/oauth/web-apps/get-user-info.html` |

## Supported Endpoints

`base` is the configured base path (default `/zoho`).

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `{base}/oauth/v2/auth` | Issues a fake authorization code and immediately redirects (302) to `redirect_uri` with `code` and the echoed `state`. |
| `POST` | `{base}/oauth/v2/token` | Exchanges a fake authorization code for an access token. |
| `GET` | `{base}/oauth/user/info` | Returns deterministic user info; requires the `Zoho-oauthtoken` auth scheme. |
| `POST` | `{base}/test/reset` | Clears local OAuth state for test isolation (loopback callers only). |

Notes on behavior:

- **Authorize** (`GET {base}/oauth/v2/auth`): requires `client_id` and a loopback `redirect_uri`. It never shows a login screen and responds with `302` to `redirect_uri`, appending the generated `code` and echoing the request `state`.
- **Token** (`POST {base}/oauth/v2/token`, `application/x-www-form-urlencoded`): on success returns `200` with `{"access_token":"<token>"}`. On failure (bad `grant_type`, or unknown/invalid `code`) it returns `{"error":"<reason>"}`. Following Zoho behavior, token-exchange failures use HTTP `200`; the client inspects the `error` field, not the status code. Authorization codes are one-time use and client-bound.
- **User info** (`GET {base}/oauth/user/info`): requires `Authorization: Zoho-oauthtoken <access_token>` (not `Bearer`). On success returns `200` with `{"Email":"<email>","Display_Name":"<name>"}` (capitalized keys). A missing token, an unknown token, or any non-`Zoho-oauthtoken` scheme returns `401`.

## Configurable User

The returned `Email` / `Display_Name` are configurable so an application can match a user by email:

- Defaults come from the `ZOHO_USER_EMAIL` / `ZOHO_USER_NAME` environment variables (falling back to `mockport@example.test` / `Mockport User`).
- A single flow can be overridden via the authorize query parameters `mock_email` and `mock_name`. The override is bound to the issued code and surfaced by the later user info call.

## Scenarios

| Scenario | Behavior |
| --- | --- |
| `oauth_success` | Default successful local OAuth workflow. |
| `invalid_code` | Forces token-exchange failure behavior for unknown or invalid codes. |
| `invalid_token` | Forces user info authentication failure (`401`). |

## Current Gaps And Tasks

| Priority | Task | Current source of truth |
| --- | --- | --- |
| P2 | Keep Zoho login UI, MFA, data-center/org routing, token refresh, scope enforcement, and full user profile fields as explicit known gaps. | `docs/site/support-matrix.md` |

## Verification

Run the adapter tests:

```bash
go test ./adapters/zohooauth
```

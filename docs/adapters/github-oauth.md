# GitHub OAuth Adapter Specification

[日本語版](github-oauth.ja.md)

This document describes the Mockport `github-oauth` adapter contract. It is not a copy of GitHub's documentation and does not claim full GitHub OAuth or REST API compatibility.

## Scope

The `github-oauth` adapter provides deterministic local behavior for selected GitHub OAuth and REST-style workflows:

- OAuth web application authorization redirect.
- OAuth access token exchange.
- Authenticated user profile lookup.
- Authenticated user email lookup.
- Authenticated user organization lookup.
- GitHub-like OAuth and REST error paths for invalid code, expired token, missing scope, redirect URI mismatch, and bad credentials.

## Base Path

Default base path:

```text
/github
```

Example config:

```yaml
adapters:
  github-oauth:
    enabled: true
    base_path: /github
    scenario: oauth_success
    fake_secret: mockport_github_client_secret
```

## Official Reference Map

Use this table to jump from Mockport's supported local surface to the closest official GitHub documentation. These links are references for behavior shape only; Mockport remains a deterministic local emulator.

| Mockport surface | Official reference |
| --- | --- |
| OAuth authorize redirect and access token exchange | `https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps` |
| Authenticated user profile | `https://docs.github.com/en/rest/users/users#get-the-authenticated-user` |
| Authenticated user emails | `https://docs.github.com/en/rest/users/emails#list-email-addresses-for-the-authenticated-user` |
| Authenticated user organizations | `https://docs.github.com/en/rest/orgs/orgs#list-organizations-for-the-authenticated-user` |
| OAuth token request errors | `https://docs.github.com/en/apps/oauth-apps/maintaining-oauth-apps/troubleshooting-oauth-app-access-token-request-errors` |

## Supported Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/github/login/oauth/authorize` | Creates a fake authorization code and redirects to `redirect_uri`. |
| `POST` | `/github/login/oauth/access_token` | Exchanges a fake code for a bearer token. |
| `GET` | `/github/user` | Returns a deterministic authenticated user profile. |
| `GET` | `/github/user/emails` | Returns deterministic authenticated user emails. |
| `GET` | `/github/user/orgs` | Returns deterministic authenticated user organizations. |

## Scenarios

| Scenario | Behavior |
| --- | --- |
| `oauth_success` | Default successful local OAuth workflow. |
| `invalid_code` | Returns token exchange failure behavior for unknown or invalid codes. |
| `expired_token` | Returns protected-resource authentication failures. |
| `scope_missing` | Returns scope-related failures for protected endpoints. |
| `redirect_uri_mismatch` | Returns token exchange failure behavior when the posted `redirect_uri` does not match the authorization request. Unsafe authorize redirect URIs are rejected independently of scenario. |

## Current Gaps And Tasks

| Priority | Task | Current source of truth |
| --- | --- | --- |
| P1 | Strengthen client contract assertions for `state`, redirect URI mismatch, invalid code, missing scope, and bad credentials. | `tasks/phase29_oauth_slack_client_evidence.md` |
| P1 | Extend REST subset contract for `/user`, `/user/emails`, and `/user/orgs` using bearer token authentication. | `contract/sdk/github-oauth-smoke.test.js` and `compat/fixtures/github/` |
| P1 | Add manifest evidence for OAuth/client contract status before considering provider-compatible promotion. | `tasks/phase26_provider_compatible_manifest_promotion.md` |
| P2 | Keep GitHub policy, repository permissions, SSO, org/enterprise enforcement, and app installation model as explicit known gaps. | `docs/site/support-matrix.md` |

## Verification

Run the adapter tests and client contract:

```bash
/usr/local/go/bin/go test ./adapters/githuboauth
bash scripts/run-sdk-contracts.sh github-oauth
```

# Zoho OAuth Example

[日本語版](README.ja.md)

This example runs the Zoho OAuth-like Mockport adapter with **fake local credentials only**. It emulates a minimal Zoho OAuth2 authorization-code flow on loopback; it is **not** full Zoho API compatibility and does not contact real Zoho.

Adapter specification: [Zoho OAuth adapter](../../docs/adapters/zoho-oauth.md)

```bash
docker build -t mockport:local -f docker/Dockerfile .
mockport run --config examples/zoho-oauth/mockport.yml
```

Use these values in the application under test:

```env
ZOHO_AUTH_BASE_URL=http://localhost:43101/zoho
ZOHO_OAUTH_CLIENT_ID=mockport_zoho_client
ZOHO_OAUTH_CLIENT_SECRET=mockport_zoho_secret
```

Optional deterministic user overrides (defaults: `mockport@example.test` / `Mockport User`):

```env
ZOHO_USER_EMAIL=mockport@example.test
ZOHO_USER_NAME=Mockport User
```

Smoke test:

```bash
REDIRECT_URL="$(curl -fsS -o /dev/null -w '%{redirect_url}' "http://localhost:43101/zoho/oauth/v2/auth?client_id=mockport_zoho_client&redirect_uri=http://localhost:3000/callback&state=local")"
CODE="$(python3 -c 'import sys, urllib.parse as u; print(u.parse_qs(u.urlparse(sys.argv[1]).query)["code"][0])' "$REDIRECT_URL")"
TOKEN="$(curl -fsS -X POST http://localhost:43101/zoho/oauth/v2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode grant_type=authorization_code \
  --data-urlencode client_id=mockport_zoho_client \
  --data-urlencode client_secret=mockport_zoho_secret \
  --data-urlencode redirect_uri=http://localhost:3000/callback \
  --data-urlencode code="$CODE" \
  | python3 -c 'import json, sys; print(json.load(sys.stdin)["access_token"])')"
curl -H "Authorization: Zoho-oauthtoken $TOKEN" http://localhost:43101/zoho/oauth/user/info
curl http://localhost:43101/_mockport/report
```

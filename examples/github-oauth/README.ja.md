# GitHub OAuth Example 日本語版

[English](README.md)

この example は、Mockport の GitHub OAuth-like adapter を使って local OAuth flow を検証するための最小構成です。

## 確認すること

- local base URL と fake client credential を使うこと。
- authorize redirect、token exchange、profile lookup の流れ。
- 実 GitHub OAuth app や production secret を使わないこと。

## Smoke test

```bash
REDIRECT_URL="$(curl -fsS -o /dev/null -w '%{redirect_url}' "http://localhost:43101/github/login/oauth/authorize?client_id=mockport_github_client&redirect_uri=http://localhost:3000/callback&state=local")"
CODE="$(python3 -c 'import sys, urllib.parse as u; print(u.parse_qs(u.urlparse(sys.argv[1]).query)["code"][0])' "$REDIRECT_URL")"
TOKEN="$(curl -fsS -X POST http://localhost:43101/github/login/oauth/access_token \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode client_id=mockport_github_client \
  --data-urlencode client_secret=mockport_github_secret \
  --data-urlencode redirect_uri=http://localhost:3000/callback \
  --data-urlencode code="$CODE" \
  | python3 -c 'import json, sys; print(json.load(sys.stdin)["access_token"])')"
curl -H "Authorization: Bearer $TOKEN" http://localhost:43101/github/user
```

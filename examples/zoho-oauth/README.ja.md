# Zoho OAuth Example 日本語版

[English](README.md)

この example は、Mockport の Zoho OAuth-like adapter を使って **fake / local 値だけ**で authorization-code flow を検証するための最小構成です。実 Zoho には接続せず、loopback 上で OAuth ログインを完結する **local emulation** です。Zoho API 全体の互換性はありません。

Adapter 仕様: [Zoho OAuth adapter 日本語版](../../docs/adapters/zoho-oauth.ja.md)

## 確認すること

- local base URL と fake client credential を使うこと。
- authorize redirect、token exchange、`Zoho-oauthtoken` scheme での user info 取得の流れ。
- 実 Zoho OAuth app や production secret を使わないこと。

## 起動

```bash
docker build -t mockport:local -f docker/Dockerfile .
mockport run --config examples/zoho-oauth/mockport.yml
```

アプリ側では次の値を向けます:

```env
ZOHO_AUTH_BASE_URL=http://localhost:43101/zoho
ZOHO_OAUTH_CLIENT_ID=mockport_zoho_client
ZOHO_OAUTH_CLIENT_SECRET=mockport_zoho_secret
```

任意の deterministic user override（未設定時は `mockport@example.test` / `Mockport User`）:

```env
ZOHO_USER_EMAIL=mockport@example.test
ZOHO_USER_NAME=Mockport User
```

## Smoke test

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
```

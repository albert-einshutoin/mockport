# Examples

[日本語版](examples.ja.md)

Run commands from the repository root.

## Stripe Checkout

[Source](../../examples/stripe-checkout/README.md)

```bash
docker build -t mockport:local -f docker/Dockerfile .
docker compose -f examples/stripe-checkout/docker-compose.yml up
curl -X POST http://localhost:43101/stripe/v1/checkout/sessions
```

## OpenAI Chat

[Source](../../examples/openai-chat/README.md)

```bash
mockport run --config examples/openai-chat/mockport.yml
curl http://localhost:43101/openai/v1/models
curl -X POST http://localhost:43101/openai/v1/chat/completions
```

## GitHub OAuth

[Source](../../examples/github-oauth/README.md)

```bash
mockport run --config examples/github-oauth/mockport.yml
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

## Slack Message

[Source](../../examples/slack-message/README.md)

```bash
mockport run --config examples/slack-message/mockport.yml
curl -X POST http://localhost:43101/slack/api/auth.test
curl -X POST http://localhost:43101/slack/api/chat.postMessage
```

## Multi-adapter

[Source](../../examples/multi-adapter/README.md)

```bash
docker build -t mockport:local -f docker/Dockerfile .
docker compose -f examples/multi-adapter/docker-compose.yml up
curl -X POST http://localhost:43101/stripe/v1/checkout/sessions
curl http://localhost:43101/openai/v1/models
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
curl -X POST http://localhost:43101/slack/api/auth.test
curl -X POST http://localhost:43101/line/v2/bot/message/push
```

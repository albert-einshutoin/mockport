#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
if [[ -z "${GO_BIN:-}" ]]; then
  if command -v go >/dev/null 2>&1; then
    GO_BIN="go"
  else
    GO_BIN="/usr/local/go/bin/go"
  fi
fi
IMAGE_TAG="mockport:local"
WORK_DIR="$(mktemp -d)"

cleanup() {
  (cd "$WORK_DIR" && docker compose -f docker-compose.mockport.yml down >/dev/null 2>&1 || true)
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

cd "$ROOT_DIR"
"$GO_BIN" build -o "$WORK_DIR/mockport" ./cmd/mockport
docker build -t "$IMAGE_TAG" -f docker/Dockerfile .

cp examples/multi-adapter/mockport.yml "$WORK_DIR/mockport.yml"
cat >"$WORK_DIR/docker-compose.mockport.yml" <<'COMPOSE'
services:
  mockport:
    image: mockport:local
    command: ["run", "--config", "/etc/mockport/mockport.yml", "--host", "0.0.0.0"]
    ports:
      - "127.0.0.1:43101:43101"
    volumes:
      - ./mockport.yml:/etc/mockport/mockport.yml
COMPOSE

cd "$WORK_DIR"
docker compose -f docker-compose.mockport.yml up -d

for _ in $(seq 1 30); do
  if curl -fsS http://localhost:43101/health >/dev/null; then
    break
  fi
  sleep 1
done

curl -fsS http://localhost:43101/health
printf '\n'
curl -fsS -X POST http://localhost:43101/stripe/v1/checkout/sessions
printf '\n'
curl -fsS http://localhost:43101/openai/v1/models
printf '\n'
github_status="$(curl -sS -o "$WORK_DIR/github-user.json" -w '%{http_code}' http://localhost:43101/github/user)"
if [[ "$github_status" != "401" ]]; then
  cat "$WORK_DIR/github-user.json" >&2
  echo "expected github unauthenticated user request to return 401, got $github_status" >&2
  exit 1
fi
cat "$WORK_DIR/github-user.json"
printf '\n'
curl -fsS -X POST http://localhost:43101/slack/api/auth.test
printf '\n'
curl -fsS -X POST http://localhost:43101/line/v2/bot/message/push
printf '\n'
zoho_status="$(curl -sS -o /dev/null -w '%{http_code}' "http://localhost:43101/zoho/oauth/v2/auth?client_id=mockport_zoho_client&redirect_uri=http://localhost/callback&state=s1")"
if [[ "$zoho_status" != "302" ]]; then
  echo "expected zoho authorize redirect to return 302, got $zoho_status" >&2
  exit 1
fi
printf '\n'
report="$(curl -fsS http://localhost:43101/_mockport/report)"
printf '%s\n' "$report"

for adapter in stripe openai github-oauth slack line zoho-oauth; do
  if ! printf '%s' "$report" | grep -q "\"name\":\"$adapter\""; then
    echo "missing adapter in report: $adapter" >&2
    exit 1
  fi
done

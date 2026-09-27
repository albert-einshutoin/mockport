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
IMAGE_TAG="mockport:smoke-$$"
WORK_DIR="$(mktemp -d)"
PROJECT="mockport-smoke-$$"
IMAGE_CREATED=false

cleanup() {
  local result=$?
  if [[ -f "$WORK_DIR/docker-compose.mockport.yml" ]]; then
    (cd "$WORK_DIR" && docker compose -p "$PROJECT" -f docker-compose.mockport.yml down >/dev/null 2>&1) || result=1
  fi
  if [[ "$IMAGE_CREATED" == true ]]; then
    docker image rm "$IMAGE_TAG" >/dev/null || result=1
  fi
  rm -rf "$WORK_DIR"
  exit "$result"
}
trap cleanup EXIT

cd "$ROOT_DIR"
SOURCE_SHA="$(git rev-parse HEAD)"
"$GO_BIN" build -o "$WORK_DIR/mockport" ./cmd/mockport
if docker image inspect "$IMAGE_TAG" >/dev/null 2>&1; then
  echo "smoke image tag already exists: $IMAGE_TAG" >&2
  exit 1
fi
docker build -t "$IMAGE_TAG" -f docker/Dockerfile .
IMAGE_CREATED=true
BUILT_IMAGE_ID="$(docker image inspect --format '{{.Id}}' "$IMAGE_TAG")"

cd "$WORK_DIR"
"$WORK_DIR/mockport" init --adapter stripe
sed "s|image: ghcr.io/albert-einshutoin/mockport:0.3.0-preview|image: $IMAGE_TAG|" docker-compose.mockport.yml > docker-compose.local.yml
if ! grep -q "image: $IMAGE_TAG" docker-compose.local.yml; then
  echo "generated Compose image could not be replaced" >&2
  exit 1
fi
mv docker-compose.local.yml docker-compose.mockport.yml
echo "source_sha=$SOURCE_SHA image=$IMAGE_TAG image_id=$BUILT_IMAGE_ID"
echo "command=docker compose -p $PROJECT -f docker-compose.mockport.yml up -d"
docker compose -p "$PROJECT" -f docker-compose.mockport.yml up -d
CONTAINER_ID="$(docker compose -p "$PROJECT" -f docker-compose.mockport.yml ps -q mockport)"
RUN_IMAGE_ID="$(docker inspect --format '{{.Image}}' "$CONTAINER_ID")"
echo "container=$CONTAINER_ID running_image_id=$RUN_IMAGE_ID"
if [[ "$BUILT_IMAGE_ID" != "$RUN_IMAGE_ID" ]]; then
  echo "running container did not use the locally built image" >&2
  exit 1
fi

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
"$WORK_DIR/mockport" report --url http://localhost:43101/_mockport/report

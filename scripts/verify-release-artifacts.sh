#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-0.2.0-preview}"
DIST_DIR="${2:-dist}"
IMAGE="${3:-}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CHECKSUM_FILE="$DIST_DIR/checksums.txt"

fail() {
  echo "release verification failed: $*" >&2
  exit 1
}

sha256_file() {
  local file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print $1}'
  else
    shasum -a 256 "$file" | awk '{print $1}'
  fi
}

require_archive() {
  local target="$1"
  local archive="$DIST_DIR/mockport_${VERSION}_${target}.tar.gz"
  [[ -f "$archive" ]] || fail "missing archive: $archive"
  grep -Fq "mockport_${VERSION}_${target}.tar.gz" "$CHECKSUM_FILE" || fail "missing checksum entry for $target"
}

verify_checksums() {
  [[ -f "$CHECKSUM_FILE" ]] || fail "missing checksums.txt"
  while read -r expected file_name; do
    [[ -n "$expected" && -n "$file_name" ]] || continue
    local archive="$DIST_DIR/$(basename "$file_name")"
    [[ -f "$archive" ]] || fail "checksum references missing file: $file_name"
    local actual
    actual="$(sha256_file "$archive")"
    [[ "$actual" == "$expected" ]] || fail "checksum mismatch for $file_name"
  done <"$CHECKSUM_FILE"
}

verify_host_binary() {
  local go_bin="${GO_BIN:-}"
  if [[ -z "$go_bin" ]]; then
    if command -v go >/dev/null 2>&1; then
      go_bin="go"
    else
      go_bin="/usr/local/go/bin/go"
    fi
  fi

  local goos goarch target archive extract_dir binary output
  goos="$("$go_bin" env GOOS)"
  goarch="$("$go_bin" env GOARCH)"
  target="${goos}_${goarch}"
  archive="$DIST_DIR/mockport_${VERSION}_${target}.tar.gz"
  [[ -f "$archive" ]] || return 0

  extract_dir="$(mktemp -d)"
  trap 'rm -rf "$extract_dir"' RETURN
  tar -C "$extract_dir" -xzf "$archive"
  binary="$extract_dir/mockport_${VERSION}_${target}/mockport"
  [[ -x "$binary" ]] || fail "host binary is not executable: $binary"
  output="$("$binary" version)"
  [[ "$output" == "mockport ${VERSION}" ]] || fail "version output = $output"
}

verify_image() {
  [[ -n "$IMAGE" ]] || return 0
  docker pull "$IMAGE" >/dev/null
  docker image inspect "$IMAGE" >/dev/null
}

cd "$ROOT_DIR"
require_archive linux_amd64
require_archive linux_arm64
require_archive darwin_amd64
require_archive darwin_arm64
verify_checksums
verify_host_binary
verify_image

echo "release artifacts verified: $VERSION"

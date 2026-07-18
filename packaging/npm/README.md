# Mockport npm Wrapper

[日本語版](README.ja.md)

This package is experimental. It is a convenience wrapper around a Mockport binary or Docker image, not the primary runtime.

Resolution order:

1. `MOCKPORT_BIN` if set.
2. A packaged binary under `vendor/<platform>/<arch>/mockport`.
3. Docker fallback using `MOCKPORT_IMAGE` if set, otherwise `ghcr.io/albert-einshutoin/mockport:0.2.0-preview`.

Use explicit release image tags for reproducible CI. The `latest` tag follows the default branch image and is not the preview release contract.

For production CI, prefer the Docker image or release binary directly.

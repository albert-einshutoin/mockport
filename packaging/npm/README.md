# Mockport npm Wrapper

[日本語版](README.ja.md)

This package is experimental. It is a convenience wrapper around a Mockport binary or Docker image, not the primary runtime.

Resolution order:

1. `MOCKPORT_BIN` if set.
2. A packaged binary under `vendor/<platform>/<arch>/mockport`.
3. Docker fallback using `ghcr.io/albert-einshutoin/mockport:latest`.

For production CI, prefer the Docker image or release binary directly.

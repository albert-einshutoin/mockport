# Docker Runtime 日本語版

[English](09_docker_runtime.md)

Mockport は Docker-first runtime を前提にしています。CI や local development で同じ emulator image を使えることを重視します。

## 方針

- container は local-only port bind を基本にします。
- config は volume mount または generated file で渡します。
- Docker Compose は application と Mockport の接続例を提供します。

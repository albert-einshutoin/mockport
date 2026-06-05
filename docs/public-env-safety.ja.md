# Public Env Safety 日本語版

[English](public-env-safety.md)

Mockport が生成する env file は、fake value のままであれば commit 可能にすることを目的にしています。

## Safe values

- `mockport_` prefix の fake credential。
- `whsec_mockport` の fake webhook signing secret。
- `http://localhost:43101` など local-only URL。

## Unsafe values

- 実 provider API key、token、webhook secret。
- production provider URL。
- 顧客 payload や captured response。
- 将来 real value に置き換わりそうな曖昧 placeholder。

確認には `bash scripts/check-public-env.sh` を使います。

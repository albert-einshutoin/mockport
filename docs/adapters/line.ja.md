# LINE Adapter 日本語版

[English](line.md)

LINE adapter は、Messaging API、LINE Login、LIFF helper、MINI App service message、LINE Pay、Mini Dapp helper の local workflow を扱います。

## 対応範囲

- message send、content、signed webhook、rich menu、channel token workflow。
- OAuth code/token/profile と local profile lookup。LINE Login flow では authorize と token exchange の `client_id` を必須とし、token exchange の値は code 発行時と一致する必要があります。
- LIFF browser runtime、provider-driven webhook redelivery、quota enforcement、regional policy、Dapp Portal の完全再現は対象外です。

詳細な endpoint と known gap は英語版を正とします。

## Verification

adapter と package test:

```bash
go test ./adapters/line ./internal/server ./internal/cli ./internal/config
```

public trust gate:

```bash
bash scripts/check-public-trust.sh
```

互換性ステータスと known gap は [`docs/site/support-matrix.md`](../site/support-matrix.md) を正とします。

SDK contract harness は LINE 未対応です。`bash scripts/run-sdk-contracts.sh line` は `unsupported provider: line` で終了します。LINE smoke test が [`contract/sdk/README.md`](../../contract/sdk/README.md) に追加されるまで、上記 adapter test を使用してください。

# OpenAI Adapter 日本語版

[English](openai.md)

共通ルール: [シナリオポリシー](../scenario-policy.ja.md)。

OpenAI adapter は、OpenAI-compatible な local API surface を使って、AI application の統合 path を secret-free に検証するための adapter です。

## 対応範囲

- models、chat completions（Chat Completions streaming を含む）、responses、embeddings、files、batches。
- deterministic fake inference と stateful response lookup。
- 実 model 品質、tokenization parity、hosted tools、provider scheduling の再現は対象外です。

`auth_required: true` を設定すると、`/openai/v1/...` は `Authorization: Bearer <fake_secret>` を厳密に検証し、欠落・不正形式・誤キーに `401` / `invalid_api_key` を返します。未設定時は従来どおり検証しません。`fake_secret` 省略時は `FakeEnv` と同じ既定のfake keyを使います。`/openai/test/reset` は引き続きloopback限定です。`auth_error` scenarioは別の強制失敗です。

詳細な request/response contract と known gap は英語版を正とします。

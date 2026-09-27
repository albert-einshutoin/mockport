# OpenAI Adapter 日本語版

[PythonアプリE2E](../../examples/app-e2e/README.md)では、HTTP入口から公式 `openai==2.46.0` SDKを経由して、このadapterへ接続します。非streaming、Chat Completions streaming、アプリ側で見える認証失敗と429、有限回の再試行、timeoutと呼び出し側切断後の上限を確認します。固定のMockport応答は接続・制御フローの証拠であり、AIの回答品質は評価しません。エラー用scenarioと遅延はアプリ外のテストセットアップで注入します。

[English](openai.md)

共通ルール: [シナリオポリシー](../scenario-policy.ja.md)。

OpenAI adapter は、OpenAI-compatible な local API surface を使って、AI application の統合 path を secret-free に検証するための adapter です。

## 対応範囲

- models、chat completions（Chat Completions streaming を含む）、responses、embeddings、files、batches。
- deterministic fake inference と stateful response lookup。
- 実 model 品質、tokenization parity、hosted tools、provider scheduling の再現は対象外です。

`auth_required: true` を設定すると、`/openai/v1/...` は `Authorization: Bearer <fake_secret>` を厳密に検証し、欠落・不正形式・誤キーに `401` / `invalid_api_key` を返します。未設定時は従来どおり検証しません。`fake_secret` 省略時は `FakeEnv` と同じ既定のfake keyを使います。`/openai/test/reset` は引き続きloopback限定です。`auth_error` scenarioは別の強制失敗です。

詳細な request/response contract と known gap は英語版を正とします。

Responses API の `stream:true` はMockportで未対応のため、`501` / `mockport_unsupported_responses_stream` を返し、状態を作成しません。OpenAI本体が非対応という意味ではありません。Chat Completions streamingとResponsesの非streamingは従来の対象範囲です。

Chat Completions の非streaming応答には固定の `created`、`finish_reason`、`usage` を含めます。streaming要求で `stream_options.include_usage: true` を指定すると、`[DONE]` の直前にusageチャンクを返します。usageの値は模擬値で、実際のトークン計測や推論結果を表しません。

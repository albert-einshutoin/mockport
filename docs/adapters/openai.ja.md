# OpenAI Adapter 日本語版

[English](openai.md)

OpenAI adapter は、OpenAI-compatible な local API surface を使って、AI application の統合 path を secret-free に検証するための adapter です。

## 対応範囲

- models、chat completions、responses、streaming、embeddings、files、batches。
- deterministic fake inference と stateful response lookup。
- 実 model 品質、tokenization parity、hosted tools、provider scheduling の再現は対象外です。

詳細な request/response contract と known gap は英語版を正とします。

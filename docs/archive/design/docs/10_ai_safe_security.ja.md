> **⚠️ アーカイブ注意 — Not maintained, may diverge, do not cite as authoritative.**
>
> 実装開始前の設計アーカイブです。**実装の正本ではありません。**
> 現行仕様は [docs/site/](../../../site/index.ja.md) を参照してください。

# AI-safe Security 日本語版

[English](10_ai_safe_security.md)

AI-safe security は、AI tool と CI に real secret を渡さず、unsafe config を検出するための設計です。

## 要点

- fake value prefix を明示します。
- real-looking secret と live provider URL を検出します。
- report と CLI output では secret 全文を出しません。

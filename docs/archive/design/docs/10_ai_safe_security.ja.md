# AI-safe Security 日本語版

[English](10_ai_safe_security.md)

AI-safe security は、AI tool と CI に real secret を渡さず、unsafe config を検出するための設計です。

## 要点

- fake value prefix を明示します。
- real-looking secret と live provider URL を検出します。
- report と CLI output では secret 全文を出しません。

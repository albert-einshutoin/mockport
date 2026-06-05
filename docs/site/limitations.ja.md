# Limitations 日本語版

[English](limitations.md)

Mockport は selected workflow の local integration test を目的にしており、provider の内部実装や未公開 behavior は再現しません。

## 対象外

- 実 payment processing、fraud、settlement、billing network。
- 実 AI inference、tokenization parity、private scheduling。
- GitHub/Slack/LINE の enterprise policy や full directory state。
- provider sandbox や production validation の完全な代替。

採用前に support matrix、report、adapter examples を確認してください。

# Compatibility Model 日本語版

[English](compatibility-model.md)

Mockport の compatibility model は、supported workflow を段階的に明示するための分類です。provider の完全 clone ではなく、local integration test に必要な再現性と known gap の透明性を重視します。

## 段階

- `experimental`: selected workflow の初期対応。
- `partial`: common workflow と unsupported behavior を文書化。
- `sdk-compatible`: selected SDK calls が local Mockport に対して通る状態。
- `workflow-compatible`: fake state、error、replayable behavior を含む状態。
- `provider-compatible`: manifest、SDK contract、fixture、known-gap report で支えられた状態。

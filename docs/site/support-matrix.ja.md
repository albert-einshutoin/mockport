# Support Matrix 日本語版

[English](support-matrix.md)

Mockport の support は explicit かつ scenario-driven です。adapter ごとに maturity、endpoint、scenario、known gap を確認してください。

## Maturity

- `experimental`: selected workflow の初期対応。
- `partial`: common workflow と unsupported behavior を文書化。
- `sdk-compatible`: selected SDK call が local Mockport に対して通る状態。
- `workflow-compatible`: fake state、error、replayable behavior を含む状態。
- `provider-compatible`: manifest、SDK contract、fixture、known-gap report で支えられた状態。

## Adapter coverage

Built-in adapter は `stripe`、`openai`、`github-oauth`、`slack`、`line` です。

SDK evidence は English support matrix と compatibility report を正とし、current SDK contract は `stripe@22.2.0` と `openai@6.39.1` です。

詳細な endpoint、scenario、known gap は英語版を正とします。

# Examples 日本語版

[English](examples.md)

## Node SDK Client

Stripe、OpenAI、LINE の公式 Node.js SDK を1つの local Mockport に接続する
実行可能な例は [`examples/node-sdk-clients/`](../../examples/node-sdk-clients/README.md)
にあります。

examples は adapter ごとの最小 integration path を確認するための入口です。各 example は fake credential と local endpoint を前提にしています。

## 対象

- Stripe checkout。
- OpenAI chat。
- GitHub OAuth。
- Slack message。
- 複数 adapter の同時起動。

実 secret は examples に含めないでください。

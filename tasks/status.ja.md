# Task Status 日本語版

[English](status.md)

この文書は task phase の進捗 overview です。正確な状態は英語版、現在の branch、CI、GitHub issue/PR を合わせて確認してください。

## 次の製品実証（2026-09-27）

P0の [PR #384](https://github.com/albert-einshutoin/mockport/pull/384)、[PR #385](https://github.com/albert-einshutoin/mockport/pull/385)、[PR #387](https://github.com/albert-einshutoin/mockport/pull/387) と、P1の [PR #389](https://github.com/albert-einshutoin/mockport/pull/389)、[PR #390](https://github.com/albert-einshutoin/mockport/pull/390) はmainへ入りました。選択したStripe／OpenAIアプリのHTTP入口から公式SDK、業務結果までのE2Eとmutation検証があります。#84 の選択 source フローは、署名付き Slack message 配送と公式SDKのthread返信を追加し、[app E2E](../examples/app-e2e/README.md)とadapterテストでSDK応答と保存状態を分けて確認します。[digest固定の試用手順](../docs/site/app-trial.ja.md)も参照してください。版付き `v0.2.0-preview` には未反映で、外部アプリ・初見利用者の試行は未確認です。

## 確認ポイント

- 完了済み phase と残 task。
- public preview、compatibility track、adapter reference docs の状態。
- 実装済みと planned の境界。

# Phase 32 Service Baseline Execution 日本語版

[English](phase32_service_baseline_execution.md)

service baseline を実行し、adapter support と report evidence を確認する phase です。

## 現在の製品gate（2026-09-27）

先にP0の [#315](https://github.com/albert-einshutoin/mockport/issues/315) checkout image確認、[#82](https://github.com/albert-einshutoin/mockport/issues/82) fake-key認証、[#386](https://github.com/albert-einshutoin/mockport/issues/386) Responses streaming未対応の明示を進めます。続くP1は [#86](https://github.com/albert-einshutoin/mockport/issues/86) のStripe／OpenAIアプリE2Eで、同じMockportプロセス、業務結果、失敗・再試行、意図的なアプリ不具合の検出をローカルとCIで証明します。これらは公開版に未反映です。P2は [#84](https://github.com/albert-einshutoin/mockport/issues/84) の署名付きSlack message eventとSDK返信です。

以下の広いbaselineは将来候補の目録です。Block Kit、LINE追加、Responses SSE、SendGridは具体的な利用者フローと検証・保守条件が揃ってから優先します。

## 確認ポイント

- この phase の正確な checklist、完了条件、検証 command は英語版を正とします。
- 実装済みか planned かは `tasks/status.md`、現在の branch、CI、GitHub issue/PR の live state と合わせて確認します。
- docs、tests、compatibility evidence に影響する変更では、関連ファイルを同時に更新します。

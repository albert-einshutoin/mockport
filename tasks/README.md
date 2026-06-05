# Mockport Task Board

[日本語版](README.ja.md)

この `tasks/` ディレクトリは、Mockport を Phase 0 から公開 OSS 運用まで TDD ベースで実装するための作業ボードです。

## 方針

- Stripe first で進める。
- Minimal MVP までは Go + Docker + built-in adapter に集中する。
- npm wrapper は後回しにする。
- Rust component は後回しにする。
- 動的 plugin system、複雑な DSL、adapter-specific image は Minimal MVP では作らない。
- Phase 11 以降は、provider 内部再現ではなく、公開 API / 公式 SDK / 主要 workflow / fake state / 主要 error shape の高忠実度互換を目標にする。
- 「Full-compatible」は全 provider 内部や全 undocumented behavior の再現ではなく、測定可能な provider-compatible local API として定義する。
- production code は、対応する failing test を先に作って失敗を確認してから実装する。

## ファイル構成

```txt
tasks/
  README.md
  status.md
  phase0_baseline.md
  phase1_stripe_minimal_mvp.md
  phase2_cli_ux.md
  phase3_ai_safe_mode.md
  phase4_trust_reports.md
  phase5_additional_adapters.md
  phase6_distribution.md
  phase7_public_oss_hardening.md
  phase8_public_env_safety.md
  phase9_public_docs_and_discovery.md
  phase10_public_preview_release.md
  phase11_community_and_maintenance.md
  phase12_fixture_spec_policy.md
  phase13_public_preview_contract_cleanup.md
  phase14_compatibility_engine.md
  phase15_sdk_contract_harness_foundation.md
  phase16_state_foundation.md
  phase17_adapter_state_adoption.md
  phase18_stripe_provider_compatibility.md
  phase19_openai_provider_compatibility.md
  phase20_github_oauth_provider_compatibility.md
  phase21_slack_provider_compatibility.md
  phase22_provider_compatible_release_track.md
  phase22_1_go_engineering_hardening.md
  phase22_2_go_10_of_10_hardening.md
  phase23_roadmap_docs_alignment.md
  phase24_github_actions_execution_recovery.md
  phase25_sdk_contract_all_provider_harness.md
  phase26_provider_compatible_manifest_promotion.md
  phase27_stripe_provider_compatible_track.md
  phase28_openai_provider_compatible_track.md
  phase29_oauth_slack_client_evidence.md
  phase30_v0_2_preview_release.md
  phase31_adapter_reference_docs.md
  phase32_service_baseline_execution.md
```

## Status の意味

```txt
pending      未着手
in_progress 進行中
blocked      外部要因または設計判断待ち
done         完了、検証済み
```

進捗の正本は `tasks/status.md` です。Phase ファイルは実行計画として使い、作業開始時に `tasks/status.md` の status を `in_progress` に変更し、検証コマンドが通ったら `done` にします。Phase ファイル内の `Status` とチェックボックスは、実行時に可能な範囲で同期します。

## TDD 作成方法

各実装タスクは次の順番で進めます。

1. RED: 期待する振る舞いを示すテストを書く。
2. RED verify: 対象テストだけを実行し、期待どおり失敗することを確認する。
3. GREEN: テストを通す最小実装を書く。
4. GREEN verify: 対象テストを再実行し、成功を確認する。
5. REFACTOR: 重複や命名を整理する。振る舞いは増やさない。
6. FULL verify: `go test ./...` など、その Phase の検証コマンドを通す。
7. Status update: `tasks/status.md` と Phase ファイルのチェックを更新する。
8. Commit checkpoint: Git リポジトリ化後は、Phase または小さな機能単位で commit する。

## Phase の境界

Phase 0 は「Mockport の実装基盤が起動して health check に答える」までです。

Phase 1 は「Stripe-like adapter を Docker で起動し、成功/失敗/認証エラー/rate limit/timeout/webhook/report/AI-safe warning を Minimal MVP の範囲で扱える」までです。

Phase 2 は「空ディレクトリから `mockport init` と `mockport up/run` で 2 分以内に起動できる CLI UX」までです。

Phase 3 は「AI-safe mode が警告、strict fail、redaction、report、docs で明確に差別化される」までです。

Phase 4 は「adapter を増やす前に、対応範囲、未対応 endpoint、scenario coverage、maturity を report で説明できる trust foundation」を作るまでです。

Phase 5 は「Phase 4 の metadata/report contract に沿って、OpenAI-compatible、GitHub OAuth-like、Slack-like の built-in adapters を追加する」までです。

Phase 6 は「GHCR、GitHub release binaries、Homebrew tap 下準備、npm wrapper 下準備、docs site の配布面を検証できる」までです。

Phase 7 は「公開前に必要な LICENSE、SECURITY、CONTRIBUTING、issue/PR templates、public CI gate を揃える」までです。

Phase 8 は「Mockport 用 `.env` を公開してもよい条件を scanner、docs、examples、CI で保証する」までです。

Phase 9 は「公開 docs、support matrix、limitations、examples を整え、初見ユーザーが導入判断できる」までです。

Phase 10 は「初回 public preview release として GitHub Release、checksums、GHCR image を実際に公開し、clean install を検証する」までです。

Phase 11 は「Dependabot、maintainer guide、roadmap、adapter contribution quality bar を整え、継続運用できる」までです。

Phase 12 は「fixture/spec snapshot/scenario の安全性、出典、更新ルールを定義する」までです。

Phase 13 は「public preview で露出した `mockport up`、OpenAI streaming、adapter helper 境界の期待値差を解消する」までです。

Phase 14 は「互換性を manifest、score、report、promotion rule として測れる Compatibility Engine を作る」までです。

Phase 15 は「公式 SDK / 実 client contract を local Mockport に向けて通す harness foundation を作る」までです。

Phase 16 は「deterministic fake state、idempotency、validation、report hook の共通基盤を作る」までです。

Phase 17 は「主要 adapter に shared state を適用し、provider-specific compatibility track の前提を作る」までです。

Phase 18 は「Stripe adapter を SDK/workflow-compatible local API に引き上げる」までです。

Phase 19 は「OpenAI adapter を SDK/workflow-compatible local API に引き上げる」までです。

Phase 20 は「GitHub OAuth adapter を client/workflow-compatible local API に引き上げる」までです。

Phase 21 は「Slack adapter を client/workflow-compatible local API に引き上げる」までです。

Phase 22 は「provider-compatible release track と compatibility report を継続運用する」までです。

Phase 22.1 は「Phase 23 の docs alignment に入る前に、Go の `net/http` streaming、typed metadata、deterministic report、helper 境界、軽量 performance cleanup を固める」までです。

Phase 22.2 は「Go の言語特性と標準 tooling を 10/10 レベルまで活かすために、typed response、context/error/runtime、state clone、static analysis、readiness evidence を固める」までです。

Phase 23 は「Phase 22 完了後の実態に合わせて roadmap、README、docs、changelog、compatibility report の期待値差をなくす」までです。

Phase 24 は「push 後に GitHub Actions run が作成されない問題を実測調査し、CI と compatibility workflow が観測可能に実行される状態へ戻す」までです。

Phase 25 は「`run-sdk-contracts.sh all` を placeholder から全 provider contract の実行入口へ昇格する」までです。

Phase 26 は「provider-compatible 昇格を versioned manifest と automated release check で制御し、主観的な maturity promotion を防ぐ」までです。

Phase 27 は「Stripe first 方針に従い、Stripe の選定 workflow を最初の provider-compatible 候補として深掘りする」までです。

Phase 28 は「OpenAI の選定 workflow について、real inference をしない前提で SDK、streaming、state、error の contract evidence を強化する」までです。

Phase 29 は「GitHub OAuth と Slack の client/SDK evidence を強化し、score と maturity をより説明可能にする」までです。

Phase 30 は「Phase 23-29 の成果を `v0.2.0-preview` として公開し、release artifact、GHCR、compatibility report、post-release smoke を検証する」までです。

Phase 31 は「登録済み adapter の公式参照 map と実装境界を棚卸しし、adapter docs と docs site の参照関係を揃える」までです。

Phase 32 は「登録済み adapter の最低 baseline を横断定義し、不足 slice と SendGrid 追加順序を実行可能な作業計画に落とす」までです。

## 参照資料

- `mockport_docs/docs/02_mvp_scope.md`
- `mockport_docs/docs/03_architecture.md`
- `mockport_docs/docs/04_go_engineering_guide.md`
- `mockport_docs/docs/06_config_spec.md`
- `mockport_docs/docs/07_adapter_design.md`
- `mockport_docs/docs/08_cli_spec.md`
- `mockport_docs/docs/09_docker_runtime.md`
- `mockport_docs/docs/10_ai_safe_security.md`
- `mockport_docs/docs/11_testing_strategy.md`
- `mockport_docs/docs/12_roadmap.md`

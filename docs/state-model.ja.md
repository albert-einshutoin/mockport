# State Model 日本語版

[English](state-model.md)

Mockport の state model は、adapter が local workflow を deterministic に再現し、list/retrieve、idempotency、webhook/report と整合するための土台です。

## 方針

- fake state は test run 中の再現性を優先します。
- concurrent request、idempotency、snapshot は race-free に扱います。
- state の永続化や provider full lifecycle の再現は selected workflow の範囲で判断します。
- 現行の workflow-compatible adapter は、metadata で宣言した stateful resource、idempotency、reset support を report に出します。

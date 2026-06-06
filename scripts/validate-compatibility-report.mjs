#!/usr/bin/env node
import fs from "node:fs";
import { fileURLToPath } from "node:url";

// 公開する互換性レポートのリリースゲート。各 adapter の promotion_eligible は、
// レポート生成時に internal/compat.CanPromote（昇格判定の単一の真実の源）で算出される。
// この validator はその結果を強制する。加えて、静的 JSON を検証するゲートである以上、
// stale な artifact や手編集で promotion_eligible だけを true に詐称した矛盾を検出するため、
// maturity が最低限要求する score / coverage / measured_level の整合性も確認する
// （provenance ガード。CanPromote の完全再現ではなく、ありえない組み合わせの拒否）。

const REQUIRED_ADAPTERS = ["stripe", "openai", "github-oauth", "slack", "line"];

// 各 maturity が満たすべき最低限の整合性。promotion_eligible が真実の源だが、
// ここで「ありえない組み合わせ」を弾いて自己申告の boolean への過信を防ぐ。
const MATURITY_FLOOR = {
  "sdk-compatible": { minScore: 40, coverage: ["sdk_coverage"] },
  "workflow-compatible": { minScore: 60, coverage: ["state_coverage", "error_coverage"] },
  "provider-compatible": {
    minScore: 80,
    coverage: ["sdk_coverage", "state_coverage", "error_coverage"],
    measuredLevel: "contract",
  },
};

function validateAdapter(adapter) {
  if (!adapter.name || !adapter.maturity || !Number.isInteger(adapter.score)) {
    throw new Error(`invalid adapter report entry: ${JSON.stringify(adapter)}`);
  }

  // 真実の源は Go の CanPromote。宣言した maturity がスコアリング規則上ありえない
  // ものなら promotion_eligible は false になり、ここで弾かれる。
  if (adapter.promotion_eligible !== true) {
    throw new Error(
      `${adapter.name} publishes maturity "${adapter.maturity}" but does not meet CanPromote (promotion_eligible is not true)`,
    );
  }

  // provenance ガード: promotion_eligible=true と矛盾する score/coverage/measured_level を
  // 拒否し、stale/手編集の JSON が boolean だけで昇格を詐称するのを防ぐ。
  const floor = MATURITY_FLOOR[adapter.maturity];
  if (floor) {
    if (adapter.score < floor.minScore) {
      throw new Error(
        `${adapter.name} claims ${adapter.maturity} but score ${adapter.score} < ${floor.minScore}`,
      );
    }
    for (const key of floor.coverage) {
      if (adapter[key] !== 100) {
        throw new Error(
          `${adapter.name} claims ${adapter.maturity} but ${key} is ${adapter[key]} (want 100)`,
        );
      }
    }
    if (floor.measuredLevel && adapter.measured_level !== floor.measuredLevel) {
      throw new Error(
        `${adapter.name} claims ${adapter.maturity} but measured_level is "${adapter.measured_level}" (want "${floor.measuredLevel}")`,
      );
    }
  }

  if (!Array.isArray(adapter.known_gaps)) {
    throw new Error(`${adapter.name} missing known_gaps array`);
  }
  if (adapter.known_gaps.length === 0) {
    throw new Error(`${adapter.name} must publish known gaps`);
  }
}

// validateReport throws on the first violation it finds.
export function validateReport(report) {
  if (!Array.isArray(report.adapters) || report.adapters.length < 5) {
    throw new Error("compatibility report must include at least five adapters");
  }
  for (const name of REQUIRED_ADAPTERS) {
    if (!report.adapters.some((adapter) => adapter.name === name)) {
      throw new Error(`compatibility report missing adapter: ${name}`);
    }
  }
  for (const adapter of report.adapters) {
    validateAdapter(adapter);
  }
}

// CLI: node validate-compatibility-report.mjs <report.json>
if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const reportPath = process.argv[2];
  if (!reportPath) {
    throw new Error("usage: validate-compatibility-report.mjs <report.json>");
  }
  const report = JSON.parse(fs.readFileSync(reportPath, "utf8"));
  validateReport(report);
}

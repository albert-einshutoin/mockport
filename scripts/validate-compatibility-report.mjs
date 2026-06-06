#!/usr/bin/env node
import fs from "node:fs";
import { fileURLToPath } from "node:url";

// Release gate for the published compatibility report. Each adapter's
// promotion_eligible is computed at report-generation time by
// internal/compat.CanPromote (the single source of truth for promotion). This
// validator enforces that result. Because it checks a static JSON file, it also
// verifies the minimal score / coverage / measured_level consistency a maturity
// requires, so a stale artifact or hand-edit that flips only promotion_eligible
// to true is caught (a provenance guard: reject impossible combinations rather
// than fully re-implement CanPromote).

const REQUIRED_ADAPTERS = ["stripe", "openai", "github-oauth", "slack", "line"];

// Minimal consistency each maturity must satisfy. promotion_eligible is the
// source of truth, but rejecting impossible combinations here prevents
// over-trusting a self-declared boolean.
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

  // The source of truth is Go's CanPromote. If the declared maturity is impossible
  // under the scoring rules, promotion_eligible is false and we stop here.
  if (adapter.promotion_eligible !== true) {
    throw new Error(
      `${adapter.name} publishes maturity "${adapter.maturity}" but does not meet CanPromote (promotion_eligible is not true)`,
    );
  }

  // Provenance guard: reject score/coverage/measured_level that contradict
  // promotion_eligible=true, so a stale or hand-edited JSON cannot fake a
  // promotion with the boolean alone.
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

import { test } from "node:test";
import assert from "node:assert/strict";
import { validateReport } from "./validate-compatibility-report.mjs";

// promotion_eligible is the true value computed by Go's CanPromote. These tests
// check that the validator enforces it and also rejects self-declarations that
// contradict it (the provenance guard).
function baseAdapter(overrides = {}) {
  return {
    name: "stripe",
    maturity: "workflow-compatible",
    measured_level: "workflow",
    score: 100,
    promotion_eligible: true,
    sdk_coverage: 100,
    state_coverage: 100,
    error_coverage: 100,
    known_gaps: ["gap"],
    ...overrides,
  };
}

// Replace the first adapter with the one under test; the rest are passing
// adapters so the required set of five is satisfied.
function reportWith(adapter) {
  const names = ["stripe", "openai", "github-oauth", "slack", "line"];
  const adapters = names.map((name, i) =>
    i === 0 ? { ...adapter, name } : baseAdapter({ name })
  );
  return { adapters };
}

test("accepts a report where every adapter is promotion-eligible and consistent", () => {
  assert.doesNotThrow(() => validateReport(reportWith(baseAdapter())));
});

test("rejects an adapter that is not promotion-eligible", () => {
  const adapter = baseAdapter({
    maturity: "provider-compatible",
    measured_level: "contract",
    promotion_eligible: false,
  });
  assert.throws(() => validateReport(reportWith(adapter)), /does not meet CanPromote/);
});

test("rejects an adapter missing the promotion_eligible field", () => {
  const adapter = baseAdapter();
  delete adapter.promotion_eligible;
  assert.throws(() => validateReport(reportWith(adapter)), /does not meet CanPromote/);
});

// Provenance guard: even with promotion_eligible=true, contradictory combinations are rejected.
test("rejects promotion_eligible=true that contradicts the maturity score floor", () => {
  const adapter = baseAdapter({ maturity: "provider-compatible", measured_level: "contract", score: 0 });
  assert.throws(() => validateReport(reportWith(adapter)), /score 0 < 80/);
});

test("rejects promotion_eligible=true with an impossible measured_level", () => {
  const adapter = baseAdapter({ maturity: "provider-compatible", measured_level: "wire", score: 100 });
  assert.throws(() => validateReport(reportWith(adapter)), /measured_level is "wire"/);
});

test("rejects workflow-compatible whose state coverage is not 100", () => {
  const adapter = baseAdapter({ state_coverage: 0 });
  assert.throws(() => validateReport(reportWith(adapter)), /state_coverage is 0/);
});

test("rejects an adapter missing known gaps", () => {
  const adapter = baseAdapter({ known_gaps: [] });
  assert.throws(() => validateReport(reportWith(adapter)), /must publish known gaps/);
});

test("rejects a report missing a required adapter", () => {
  const report = { adapters: [baseAdapter({ name: "stripe" }), baseAdapter({ name: "openai" })] };
  assert.throws(() => validateReport(report), /at least five adapters|missing adapter/);
});

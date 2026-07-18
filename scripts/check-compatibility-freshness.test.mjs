import assert from "node:assert/strict";
import test from "node:test";

import { assertReportFreshness } from "./check-compatibility-freshness.mjs";

test("accepts a report generated on the latest source date", () => {
  assert.doesNotThrow(() => assertReportFreshness("2026-07-19", "2026-07-19"));
});

test("accepts a report newer than its compatibility sources", () => {
  assert.doesNotThrow(() => assertReportFreshness("2026-07-20", "2026-07-19"));
});

test("rejects a report older than its compatibility sources", () => {
  assert.throws(
    () => assertReportFreshness("2026-07-02", "2026-07-16"),
    /older than compatibility source date/,
  );
});

test("rejects malformed dates", () => {
  assert.throws(() => assertReportFreshness("July 19", "2026-07-19"), /YYYY-MM-DD/);
  assert.throws(() => assertReportFreshness("2026-07-19", "latest"), /YYYY-MM-DD/);
});

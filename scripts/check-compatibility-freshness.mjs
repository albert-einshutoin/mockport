#!/usr/bin/env node
import fs from "node:fs";
import { pathToFileURL } from "node:url";

function parseISODate(value, label) {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) {
    throw new Error(`${label} must use YYYY-MM-DD format: ${value}`);
  }
  const parsed = new Date(`${value}T00:00:00Z`);
  if (Number.isNaN(parsed.valueOf()) || parsed.toISOString().slice(0, 10) !== value) {
    throw new Error(`${label} must use a valid YYYY-MM-DD date: ${value}`);
  }
  return parsed;
}

export function assertReportFreshness(reportDate, latestSourceDate) {
  const report = parseISODate(reportDate, "report generated_at");
  const source = parseISODate(latestSourceDate, "compatibility source date");
  if (report < source) {
    throw new Error(
      `report generated_at ${reportDate} is older than compatibility source date ${latestSourceDate}`,
    );
  }
}

function main() {
  const [reportPath, latestSourceDate] = process.argv.slice(2);
  if (!reportPath || !latestSourceDate) {
    throw new Error(
      "usage: check-compatibility-freshness.mjs <report.json> <latest-source-date>",
    );
  }
  const report = JSON.parse(fs.readFileSync(reportPath, "utf8"));
  if (typeof report.generated_at !== "string") {
    throw new Error(`${reportPath} must contain a string generated_at field`);
  }
  assertReportFreshness(report.generated_at, latestSourceDate);
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  main();
}

#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const [inputPath, outDir, reportDate] = process.argv.slice(2);
if (!inputPath || !outDir || !reportDate) {
  throw new Error("usage: render-compatibility-report.mjs <runtime-report.json> <out-dir> <date>");
}

const snapshot = JSON.parse(fs.readFileSync(inputPath, "utf8"));
const maturityByAdapter = new Map((snapshot.adapters || []).map((adapter) => [adapter.name, adapter.maturity]));
const stateByAdapter = new Map((snapshot.state_coverage || []).map((entry) => [entry.adapter, entry]));
const behaviorByAdapter = new Map();
for (const entry of snapshot.behavior_matrix || []) {
  if (!behaviorByAdapter.has(entry.adapter)) {
    behaviorByAdapter.set(entry.adapter, []);
  }
  behaviorByAdapter.get(entry.adapter).push(`${entry.method} ${entry.path}`);
}

const knownGaps = {
  stripe: [
    "No fraud, payment network, tax, disputes, Connect, or full Billing lifecycle.",
  ],
  openai: [
    "No real model quality, tokenization parity, hosted tools, vector stores, or provider scheduling.",
  ],
  "github-oauth": [
    "No real GitHub policy, repository permissions, SSO, org/enterprise enforcement, or app installation model.",
  ],
  slack: [
    "No real delivery, Events API completeness, Block Kit validation, files, app scopes, enterprise policy, or full workspace directory.",
  ],
  line: [
    "No official LINE SDK contract yet, no real LIFF browser runtime, no provider-driven webhook redelivery, no monthly quota/rate bucket enforcement, no complete Messaging API schema validation, no regional policy enforcement, and Mini Dapp endpoints are local SDK helpers rather than a full Dapp Portal clone.",
  ],
};

const adapters = (snapshot.compatibility || []).map((entry) => {
  const state = stateByAdapter.get(entry.adapter) || {};
  return {
    name: entry.adapter,
    maturity: maturityByAdapter.get(entry.adapter) || "experimental",
    measured_level: entry.level,
    score: entry.score,
    provider_version: entry.provider_version,
    sdk_versions: entry.sdk_versions || [],
    client_evidence: entry.client_evidence || [],
    endpoint_coverage: entry.endpoint_coverage || 0,
    scenario_coverage: entry.scenario_coverage || 0,
    sdk_coverage: entry.sdk_coverage || 0,
    state_coverage: entry.state_coverage || 0,
    error_coverage: entry.error_coverage || 0,
    stateful_resources: state.stateful_resources || [],
    idempotency: Boolean(state.idempotency),
    reset: Boolean(state.reset),
    endpoints: behaviorByAdapter.get(entry.adapter) || [],
    known_gaps: knownGaps[entry.adapter] || [],
  };
}).sort((a, b) => a.name.localeCompare(b.name));

const report = {
  generated_by: "scripts/generate-compatibility-report.sh",
  generated_at: reportDate,
  source: "_mockport/report",
  release_labels: {
    experimental: "Early adapter coverage for selected workflows. Expect gaps.",
    "sdk-compatible": "Selected SDK or client contract calls pass against local Mockport.",
    "workflow-compatible": "Selected workflows include fake state, errors, and replayable behavior.",
    "provider-compatible": "Selected provider workflows are backed by manifests, SDK contracts, fixtures, scores, and known-gap reports.",
  },
  promotion_criteria: {
    "sdk-compatible": "SDK/client contract coverage exists and score is at least 40.",
    "workflow-compatible": "Workflow, state, and error evidence exists and score is at least 60.",
    "provider-compatible": "Contract-level evidence exists and score is at least 80.",
  },
  adapters,
};

fs.mkdirSync(outDir, { recursive: true });
fs.writeFileSync(path.join(outDir, "latest.json"), `${JSON.stringify(report, null, 2)}\n`);
fs.writeFileSync(path.join(outDir, "latest.md"), renderMarkdown(report));

function renderMarkdown(report) {
  const lines = [];
  lines.push("# Compatibility Report");
  lines.push("");
  lines.push("[日本語版](latest.ja.md)");
  lines.push("");
  lines.push(`Generated: ${report.generated_at}`);
  lines.push("");
  lines.push("Compatibility is measured from Mockport runtime metadata, SDK/client contract checks, fixture coverage, and known gaps. It is not a claim that provider internals or undocumented behavior are reproduced.");
  lines.push("");
  lines.push("## Scores");
  lines.push("");
  lines.push("| Adapter | Maturity | Score | Provider API | SDK/client evidence |");
  lines.push("| --- | --- | ---: | --- | --- |");
  for (const adapter of report.adapters) {
    const sdk = evidenceLabel(adapter);
    lines.push(`| \`${adapter.name}\` | \`${adapter.maturity}\` | ${adapter.score} | ${adapter.provider_version} | ${sdk} |`);
  }
  lines.push("");
  lines.push("## Coverage");
  lines.push("");
  lines.push("| Adapter | Endpoint | Scenario | SDK/client | State | Error |");
  lines.push("| --- | ---: | ---: | ---: | ---: | ---: |");
  for (const adapter of report.adapters) {
    lines.push(`| \`${adapter.name}\` | ${adapter.endpoint_coverage} | ${adapter.scenario_coverage} | ${adapter.sdk_coverage} | ${adapter.state_coverage} | ${adapter.error_coverage} |`);
  }
  lines.push("");
  lines.push("## Known Gaps");
  lines.push("");
  for (const adapter of report.adapters) {
    lines.push(`### ${adapter.name}`);
    for (const gap of adapter.known_gaps) {
      lines.push(`- ${gap}`);
    }
    lines.push("");
  }
  lines.push("## Release Labels");
  lines.push("");
  for (const [label, description] of Object.entries(report.release_labels)) {
    lines.push(`- \`${label}\`: ${description}`);
  }
  return `${lines.join("\n")}\n`;
}

function evidenceLabel(adapter) {
  const evidence = [
    ...(adapter.sdk_versions || []),
    ...(adapter.client_evidence || []),
  ];
  return evidence.length > 0 ? evidence.join(", ") : "none";
}

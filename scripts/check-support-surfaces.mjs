#!/usr/bin/env node
import fs from "node:fs";

function read(path) {
  return fs.readFileSync(path, "utf8");
}

function uniqueSorted(values) {
  return [...new Set(values)].sort();
}

function assertSameSet(label, got, want) {
  const gotSorted = uniqueSorted(got);
  const wantSorted = uniqueSorted(want);
  if (JSON.stringify(gotSorted) !== JSON.stringify(wantSorted)) {
    throw new Error(`${label} = ${JSON.stringify(gotSorted)}, want ${JSON.stringify(wantSorted)}`);
  }
}

function requireText(path, content, text) {
  if (!content.includes(text)) {
    throw new Error(`${path} missing ${JSON.stringify(text)}`);
  }
}

function parseBuiltinAdapterNames() {
  const builtinPath = "internal/cli/builtin.go";
  const builtin = read(builtinPath);
  const imports = new Map();
  for (const match of builtin.matchAll(/"github\.com\/albert-einshutoin\/mockport\/adapters\/([^"]+)"/g)) {
    imports.set(match[1], match[1]);
  }

  const names = [];
  for (const match of builtin.matchAll(/\b([A-Za-z_][A-Za-z0-9_]*)\.New\(\)/g)) {
    const packageName = match[1];
    const adapterDir = imports.get(packageName);
    if (!adapterDir) {
      throw new Error(`${builtinPath} registers ${packageName}.New() without an adapter import`);
    }
    const adapterSource = read(`adapters/${adapterDir}/adapter.go`);
    const nameMatch = adapterSource.match(/func \(a Adapter\) Name\(\) string\s*{\s*return "([^"]+)"/);
    if (!nameMatch) {
      throw new Error(`adapters/${adapterDir}/adapter.go missing literal Adapter.Name()`);
    }
    names.push(nameMatch[1]);
  }

  if (names.length === 0) {
    throw new Error(`${builtinPath} has no built-in adapters`);
  }
  return uniqueSorted(names);
}

function parseBugTemplateAdapterOptions() {
  const path = ".github/ISSUE_TEMPLATE/bug_report.yml";
  const lines = read(path).split("\n");
  const idIndex = lines.findIndex((line) => line.trim() === "id: adapter");
  if (idIndex === -1) {
    throw new Error(`${path} missing adapter dropdown`);
  }
  const optionsIndex = lines.findIndex((line, index) => index > idIndex && line.trim() === "options:");
  if (optionsIndex === -1) {
    throw new Error(`${path} missing adapter options`);
  }

  const options = [];
  for (const line of lines.slice(optionsIndex + 1)) {
    if (/^\s{4,}\w/.test(line) && !line.includes("- ")) {
      break;
    }
    const match = line.match(/^\s*-\s+([a-z0-9-]+)\s*$/);
    if (match) {
      options.push(match[1]);
    }
  }
  return options;
}

function parseFeatureTemplateAdapterPlaceholder() {
  const path = ".github/ISSUE_TEMPLATE/feature_request.yml";
  const lines = read(path).split("\n");
  const idIndex = lines.findIndex((line) => line.trim() === "id: target_adapter");
  if (idIndex === -1) {
    throw new Error(`${path} missing target_adapter input`);
  }
  const placeholderLine = lines
    .slice(idIndex)
    .find((line) => line.trim().startsWith("placeholder:"));
  const match = placeholderLine?.match(/placeholder:\s*"([^"]+)"/);
  if (!match) {
    throw new Error(`${path} missing target_adapter placeholder`);
  }
  return match[1];
}

function parseSupportMatrixAdapters(content) {
  return uniqueSorted([...content.matchAll(/^\| `([^`]+)` \|/gm)].map((match) => match[1]));
}

function splitPackageVersion(value) {
  const index = value.lastIndexOf("@");
  if (index <= 0) {
    throw new Error(`invalid SDK version evidence: ${value}`);
  }
  return [value.slice(0, index), value.slice(index + 1)];
}

const builtInAdapters = parseBuiltinAdapterNames();
const supportMatrixPath = "docs/site/support-matrix.md";
const supportMatrix = read(supportMatrixPath);
const supportMatrixJaPath = "docs/site/support-matrix.ja.md";
const supportMatrixJa = read(supportMatrixJaPath);
const bugOptions = parseBugTemplateAdapterOptions();
const featureTemplatePath = ".github/ISSUE_TEMPLATE/feature_request.yml";
const featurePlaceholder = parseFeatureTemplateAdapterPlaceholder();
const reportPath = "docs/compatibility-reports/latest.json";
const report = JSON.parse(read(reportPath));
const packageJson = JSON.parse(read("contract/sdk/package.json"));

assertSameSet("compatibility report adapters", report.adapters.map((adapter) => adapter.name), builtInAdapters);
assertSameSet(
  "support matrix adapters",
  parseSupportMatrixAdapters(supportMatrix).filter((name) => builtInAdapters.includes(name)),
  builtInAdapters,
);

for (const name of builtInAdapters) {
  if (!bugOptions.includes(name)) {
    throw new Error(`bug_report.yml adapter dropdown missing ${name}`);
  }
  if (!featurePlaceholder.split(/,\s*|\s+or\s+/).includes(name)) {
    throw new Error(`${featureTemplatePath} target_adapter placeholder missing ${name}`);
  }
  requireText(supportMatrixJaPath, supportMatrixJa, `\`${name}\``);
}

for (const adapter of report.adapters) {
  for (const evidence of adapter.sdk_versions || []) {
    const [packageName, version] = splitPackageVersion(evidence);
    if (packageJson.devDependencies?.[packageName] !== version) {
      throw new Error(
        `${reportPath} reports ${evidence}, but contract/sdk/package.json has ${packageName}@${packageJson.devDependencies?.[packageName]}`,
      );
    }
    requireText(supportMatrixPath, supportMatrix, evidence);
    requireText(supportMatrixJaPath, supportMatrixJa, evidence);
  }
}

#!/usr/bin/env node
"use strict";

const { runSmokePlaceholder } = require("./smoke-placeholder.test.js");
const { runStripeSmoke } = require("./stripe-smoke.test.js");
const { runOpenAISmoke } = require("./openai-smoke.test.js");
const { runGitHubOAuthSmoke } = require("./github-oauth-smoke.test.js");
const { runSlackSmoke } = require("./slack-smoke.test.js");

const allowedProviders = new Set(["all", "stripe", "openai", "github-oauth", "slack"]);

const liveSmokeRunners = {
  stripe: runStripeSmoke,
  openai: runOpenAISmoke,
  "github-oauth": runGitHubOAuthSmoke,
  slack: runSlackSmoke,
};

const allProviders = ["stripe", "openai", "github-oauth", "slack"];

function parseArgs(argv) {
  const options = {
    baseURL: process.env.MOCKPORT_BASE_URL || "http://127.0.0.1:43101",
    provider: process.env.MOCKPORT_PROVIDER || "all",
    offline: false,
    json: false,
  };

  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--offline") {
      options.offline = true;
    } else if (arg === "--json") {
      options.json = true;
    } else if (arg === "--provider") {
      options.provider = argv[index + 1];
      index += 1;
    } else if (arg.startsWith("--provider=")) {
      options.provider = arg.slice("--provider=".length);
    } else if (arg === "--base-url") {
      options.baseURL = argv[index + 1];
      index += 1;
    } else if (arg.startsWith("--base-url=")) {
      options.baseURL = arg.slice("--base-url=".length);
    } else {
      throw new Error(`unknown argument: ${arg}`);
    }
  }

  if (!allowedProviders.has(options.provider)) {
    throw new Error(`unsupported provider: ${options.provider}`);
  }
  return options;
}

async function runAllSmoke(options) {
  const providers = [];
  let failed = false;
  for (const provider of allProviders) {
    const runner = liveSmokeRunners[provider];
    try {
      const result = await runner({ ...options, provider });
      providers.push(result);
    } catch (error) {
      failed = true;
      providers.push({
        provider,
        baseURL: options.baseURL,
        status: "failed",
        error: error.message,
      });
    }
  }
  return {
    provider: "all",
    baseURL: options.baseURL,
    status: failed ? "failed" : "sdk-ok",
    providers,
  };
}

function formatResultLine(result) {
  if (result.provider === "all" && Array.isArray(result.providers)) {
    const details = result.providers
      .map((entry) => `${entry.provider}=${entry.status}`)
      .join(" ");
    return `sdk-contracts provider=all status=${result.status} baseURL=${result.baseURL} ${details}\n`;
  }
  return `sdk-contracts provider=${result.provider} status=${result.status} baseURL=${result.baseURL}\n`;
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  let result;
  if (options.offline) {
    result = await runSmokePlaceholder(options);
  } else if (options.provider === "all") {
    result = await runAllSmoke(options);
  } else {
    const runner = liveSmokeRunners[options.provider];
    result = await runner(options);
  }
  if (options.json) {
    process.stdout.write(`${JSON.stringify(result)}\n`);
  } else {
    process.stdout.write(formatResultLine(result));
  }
  if (result.status === "failed") {
    process.exit(1);
  }
}

main().catch((error) => {
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
});

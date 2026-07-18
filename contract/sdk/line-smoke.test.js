"use strict";

const fs = require("node:fs");
const path = require("node:path");
const { messagingApi } = require("@line/bot-sdk");

function lineSDKLabel() {
  const packagePath = path.resolve(
    path.dirname(require.resolve("@line/bot-sdk")),
    "../../package.json",
  );
  const packageData = JSON.parse(fs.readFileSync(packagePath, "utf8"));
  return `@line/bot-sdk@${packageData.version}`;
}

async function runLINESmoke(options) {
  const client = new messagingApi.MessagingApiClient({
    channelAccessToken: "mockport_line_token",
    baseURL: options.baseURL,
  });

  const response = await client.pushMessage({
    to: "U_MOCKPORT",
    messages: [{ type: "text", text: "hello from LINE SDK" }],
  });

  if (!Array.isArray(response.sentMessages) || response.sentMessages.length !== 1) {
    throw new Error(`LINE push response missing sentMessages: ${JSON.stringify(response)}`);
  }
  if (!response.sentMessages[0].id) {
    throw new Error(`LINE push response missing message id: ${JSON.stringify(response)}`);
  }

  return {
    provider: "line",
    baseURL: options.baseURL,
    status: "sdk-ok",
    sdk: lineSDKLabel(),
    message: response.sentMessages[0].id,
  };
}

module.exports = { runLINESmoke };

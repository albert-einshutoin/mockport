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

  const reply = await client.replyMessage({
    replyToken: "mockport_reply_token",
    messages: [{ type: "text", text: "reply from LINE SDK" }],
  });
  assertObject(reply, "LINE reply response");

  const profile = await client.getProfile("Umockport");
  assertEqual(profile.userId, "Umockport", "LINE profile user id");
  assertEqual(profile.displayName, "Mockport LINE User", "LINE profile display name");

  const richMenu = await client.createRichMenu({
    size: { width: 2500, height: 1686 },
    selected: false,
    name: "Mockport SDK menu",
    chatBarText: "Open",
    areas: [],
  });
  if (!richMenu.richMenuId) {
    throw new Error(`LINE rich menu response missing id: ${JSON.stringify(richMenu)}`);
  }
  const richMenus = await client.getRichMenuList();
  if (!richMenus.richmenus.some((entry) => entry.richMenuId === richMenu.richMenuId)) {
    throw new Error(`LINE rich menu list missing ${richMenu.richMenuId}`);
  }
  await client.deleteRichMenu(richMenu.richMenuId);

  const oauth = await runOAuthContract(options.baseURL);
  await assertInvalidRequestContract(options.baseURL);

  return {
    provider: "line",
    baseURL: options.baseURL,
    status: "sdk-ok",
    sdk: lineSDKLabel(),
    message: response.sentMessages[0].id,
    profile: profile.userId,
    richMenu: richMenu.richMenuId,
    oauthToken: oauth.access_token,
  };
}

async function runOAuthContract(baseURL) {
  const authorizeURL = new URL("/line/oauth2/v2.1/authorize", baseURL);
  authorizeURL.searchParams.set("client_id", "mockport_line_channel");
  authorizeURL.searchParams.set("redirect_uri", "http://localhost/callback");
  authorizeURL.searchParams.set("state", "mockport-state");
  authorizeURL.searchParams.set("scope", "profile openid");
  const authorize = await fetch(authorizeURL, { redirect: "manual" });
  assertEqual(authorize.status, 302, "LINE authorize status");
  const location = authorize.headers.get("location");
  if (!location) {
    throw new Error("LINE authorize response missing location");
  }
  const redirect = new URL(location);
  const code = redirect.searchParams.get("code");
  if (!code) {
    throw new Error("LINE authorize response missing code");
  }

  const token = await fetch(new URL("/line/oauth2/v2.1/token", baseURL), {
    method: "POST",
    headers: { "content-type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      grant_type: "authorization_code",
      code,
      client_id: "mockport_line_channel",
      redirect_uri: "http://localhost/callback",
    }),
  });
  assertEqual(token.status, 200, "LINE token status");
  const body = await token.json();
  assertEqual(body.token_type, "Bearer", "LINE token type");
  if (!body.access_token) {
    throw new Error(`LINE token response missing access_token: ${JSON.stringify(body)}`);
  }
  return body;
}

async function assertInvalidRequestContract(baseURL) {
  const response = await fetch(new URL("/v2/bot/message/push", baseURL), {
    method: "POST",
    headers: {
      "authorization": "Bearer mockport_line_token",
      "content-type": "application/json",
      "x-mockport-scenario": "invalid_request",
    },
    body: JSON.stringify({ to: "Umockport", messages: [] }),
  });
  assertEqual(response.status, 400, "LINE invalid request status");
  const body = await response.json();
  if (!body.message) {
    throw new Error(`LINE invalid request missing message: ${JSON.stringify(body)}`);
  }
}

function assertEqual(got, want, label) {
  if (got !== want) {
    throw new Error(`${label}: got ${got}, want ${want}`);
  }
}

function assertObject(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label}: got ${JSON.stringify(value)}, want object`);
  }
}

module.exports = { runLINESmoke };

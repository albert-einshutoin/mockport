"use strict";

async function runSlackSmoke(options) {
  const baseURL = new URL("/slack/api/", options.baseURL).toString();
  const token = "mockport_slack_token";

  const auth = await slackAPI(baseURL, "auth.test", token, {});
  assertEqual(auth.ok, true, "auth ok");
  assertEqual(auth.user_id, "U_MOCKPORT", "auth user");
  assertEqual(auth.team_id, "T_MOCKPORT", "auth team id");

  const channels = await slackAPI(baseURL, "conversations.list", token, {});
  assertEqual(channels.ok, true, "channels ok");
  assertEqual(channels.channels[0].id, "C_MOCKPORT", "default channel id");

  const posted = await slackAPI(baseURL, "chat.postMessage", token, {
    channel: "C_MOCKPORT",
    text: "hello from client",
  });
  assertEqual(posted.ok, true, "post ok");
  assertEqual(posted.message.text, "hello from client", "post text");

  const updated = await slackAPI(baseURL, "chat.update", token, {
    channel: "C_MOCKPORT",
    ts: posted.ts,
    text: "edited from client",
  });
  assertEqual(updated.ok, true, "update ok");
  assertEqual(updated.message.text, "edited from client", "update text");

  const history = await slackAPI(baseURL, "conversations.history", token, {
    channel: "C_MOCKPORT",
  });
  assertEqual(history.ok, true, "history ok");
  assertEqual(history.messages[0].text, "edited from client", "history text");

  const deleted = await slackAPI(baseURL, "chat.delete", token, {
    channel: "C_MOCKPORT",
    ts: posted.ts,
  });
  assertEqual(deleted.ok, true, "delete ok");

  return {
    provider: "slack",
    baseURL: options.baseURL,
    status: "client-ok",
    team: auth.team_id,
    channel: posted.channel,
    message: posted.ts,
  };
}

async function slackAPIRequest(baseURL, method, token, params) {
  const response = await fetch(new URL(method, baseURL), {
    method: "POST",
    headers: {
      "authorization": `Bearer ${token}`,
      "content-type": "application/x-www-form-urlencoded",
    },
    body: new URLSearchParams(params),
  });
  const body = await response.json();
  return { response, body };
}

async function slackAPI(baseURL, method, token, params) {
  const { response, body } = await slackAPIRequest(baseURL, method, token, params);
  if (!response.ok || body.ok !== true) {
    throw new Error(`${method} failed: status=${response.status} body=${JSON.stringify(body)}`);
  }
  return body;
}

async function runSlackRateLimitedContract(options) {
  const baseURL = new URL("/slack/api/", options.baseURL).toString();
  const token = "mockport_slack_token";

  const { response, body } = await slackAPIRequest(baseURL, "chat.postMessage", token, {
    channel: "C_MOCKPORT",
    text: "hello from client",
  });

  assertEqual(response.status, 429, "rate limit status");
  assertEqual(response.headers.get("Retry-After"), "1", "Retry-After header");
  assertEqual(body.ok, false, "rate limit ok");
  assertEqual(body.error, "ratelimited", "rate limit error");

  return {
    provider: "slack",
    baseURL: options.baseURL,
    status: "client-ok",
    contract: "rate_limited",
  };
}

function assertEqual(got, want, label) {
  if (got !== want) {
    throw new Error(`${label}: got ${got}, want ${want}`);
  }
}

module.exports = { runSlackSmoke, runSlackRateLimitedContract };

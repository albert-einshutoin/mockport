import { messagingApi } from "@line/bot-sdk";

const mockportURL = process.env.MOCKPORT_BASE_URL || "http://127.0.0.1:43101";
const client = new messagingApi.MessagingApiClient({
  channelAccessToken: "mockport_line_token",
  baseURL: mockportURL,
});

// The official SDK sends absolute /v2/bot paths, so Mockport exposes a
// narrowly scoped SDK-compatible alias in addition to /line/v2/bot routes.
const response = await client.pushMessage({
  to: "U_MOCKPORT",
  messages: [{ type: "text", text: "Hello from the Mockport example" }],
});

console.log(JSON.stringify(response, null, 2));

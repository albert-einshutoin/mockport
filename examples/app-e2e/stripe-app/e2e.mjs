import assert from "node:assert/strict";
import { createHmac } from "node:crypto";

const mockport = process.env.MOCKPORT_BASE_URL || "http://127.0.0.1:43101";
const app = process.env.APP_BASE_URL || "http://127.0.0.1:33001";

async function json(url, options = {}) {
  const response = await fetch(url, { ...options, signal: options.signal ?? AbortSignal.timeout(5000) });
  const text = await response.text();
  let body;
  try { body = JSON.parse(text); } catch { throw new Error(`${url} returned ${response.status}: ${text}`); }
  return { status: response.status, body };
}
function post(url, body, headers = {}) {
  return json(url, { method: "POST", headers: { "Content-Type": "application/json", ...headers }, body: JSON.stringify(body) });
}
async function waitFor(url) {
  for (let attempt = 0; attempt < 30; attempt++) {
    try {
      if ((await json(url, { signal: AbortSignal.timeout(500) })).status === 200) return;
    } catch { /* The Compose service may still be starting. */ }
    await new Promise((resolve) => setTimeout(resolve, 200));
  }
  throw new Error(`${url} did not become ready`);
}
async function order(id) {
  const created = await post(`${app}/orders`, { order_id: id });
  assert.equal(created.status, 201, JSON.stringify(created));
  assert.equal(created.body.status, "pending");
  const retrieved = await json(`${app}/orders/${id}/session`);
  assert.equal(retrieved.status, 200);
  assert.equal(retrieved.body.id, created.body.session_id);
  assert.equal(retrieved.body.client_reference_id, id);
  assert.equal(retrieved.body.payment_status, "unpaid");
  return created.body.session_id;
}
async function state(id, status, updates, events) {
  const result = await json(`${app}/orders/${id}`);
  assert.equal(result.status, 200);
  assert.equal(result.body.status, status, JSON.stringify(result));
  assert.equal(result.body.paid_updates, updates, JSON.stringify(result));
  assert.equal(result.body.processed_events, events, JSON.stringify(result));
}
async function send(sessionID, eventID, eventType) {
  const result = await post(`${mockport}/stripe/test/webhook/send`, {
    session_id: sessionID, event_id: eventID, event_type: eventType,
  });
  assert.equal(result.status, 202, JSON.stringify(result));
  assert.equal(result.body.event_id, eventID);
}

await waitFor(`${mockport}/health`);
await waitFor(`${app}/health`);
const invalidJSON = await json(`${app}/orders`, { method: "POST", headers: { "Content-Type": "application/json" }, body: "{" });
assert.equal(invalidJSON.status, 400);
assert.equal(invalidJSON.body.error, "invalid_json");
const oversized = await json(`${app}/orders`, { method: "POST", headers: { "Content-Type": "application/json" }, body: "x".repeat(64 * 1024 + 1) });
assert.equal(oversized.status, 413);
assert.equal(oversized.body.error, "request_too_large");
const first = await order("order_1");
const second = await order("order_2");
const failed = await order("order_3");
await state("order_1", "pending", 0, 0);
await send(first, "evt_order_1", "checkout.session.completed");
await state("order_1", "paid", 1, 1);
await send(first, "evt_order_1", "checkout.session.completed");
await state("order_1", "paid", 1, 1);
await send(second, "evt_order_2", "checkout.session.completed");
await state("order_2", "paid", 1, 1);
await send(failed, "evt_order_3_failed", "checkout.session.async_payment_failed");
await state("order_3", "failed", 0, 1);

for (const signature of ["bad-signature", undefined]) {
  const bad = await post(`${app}/webhooks/stripe`, {
    id: "evt_bad", type: "checkout.session.completed",
    data: { object: { id: failed, object: "checkout.session", client_reference_id: "order_3", payment_status: "paid" } },
  }, signature ? { "Stripe-Signature": signature } : {});
  assert.equal(bad.status, 400);
  assert.equal(bad.body.error, "invalid_signature");
}
const signed = JSON.stringify({
  id: "evt_tampered", type: "checkout.session.completed",
  data: { object: { id: failed, object: "checkout.session", client_reference_id: "order_3", payment_status: "unpaid" } },
});
const timestamp = Math.floor(Date.now() / 1000);
const digest = createHmac("sha256", "whsec_mockport").update(`${timestamp}.${signed}`).digest("hex");
const tampered = await json(`${app}/webhooks/stripe`, {
  method: "POST",
  headers: { "Content-Type": "application/json", "Stripe-Signature": `t=${timestamp},v1=${digest}` },
  body: signed.replace('"unpaid"', '"paid"'),
});
assert.equal(tampered.status, 400);
assert.equal(tampered.body.error, "invalid_signature");
await state("order_3", "failed", 0, 1);

const missing = await post(`${mockport}/stripe/test/webhook/send`, {
  session_id: "cs_missing", event_id: "evt_missing", event_type: "checkout.session.completed",
});
assert.equal(missing.status, 404);
const report = await json(`${mockport}/_mockport/report`);
assert.equal(report.status, 200);
assert.equal(report.body.request_history.truncated, false);
assert.equal(report.body.adapters.find(({ name }) => name === "stripe")?.auth_required, true);
assert.ok(report.body.requests.some(({ path, status }) => path === "/stripe/test/webhook/send" && status === 202));
console.log(JSON.stringify({ passed: true, sessions: [first, second, failed], webhook_deliveries: 4, duplicate_paid_updates: 1, rejected_signatures: 3, request_history: report.body.request_history }));

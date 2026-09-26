import { createServer } from "node:http";
import Stripe from "stripe";

const baseURL = new URL(process.env.MOCKPORT_BASE_URL);
const apiKey = process.env.STRIPE_SECRET_KEY;
const webhookSecret = process.env.STRIPE_WEBHOOK_SECRET;
if (!apiKey || !webhookSecret || !["127.0.0.1", "localhost", "mockport"].includes(baseURL.hostname)) {
  throw new Error("Set fake Stripe keys and a local Mockport URL before starting the example");
}
const stripe = new Stripe(apiKey, {
  apiVersion: "2025-10-29.clover",
  host: baseURL.hostname,
  port: Number(baseURL.port),
  protocol: baseURL.protocol.replace(":", ""),
  maxNetworkRetries: 0,
  telemetry: false,
});
const orders = new Map();
const processedEvents = new Set();
class RequestTooLargeError extends Error {}

function reply(response, status, body) {
  response.writeHead(status, { "Content-Type": "application/json" });
  response.end(JSON.stringify(body));
}

async function bodyBytes(request) {
  const chunks = [];
  let size = 0;
  for await (const chunk of request) {
    size += chunk.length;
    if (size > 64 * 1024) throw new RequestTooLargeError("request body too large");
    chunks.push(chunk);
  }
  return Buffer.concat(chunks);
}

const server = createServer(async (request, response) => {
  try {
    if (request.method === "GET" && request.url === "/health") {
      return reply(response, 200, { ready: true });
    }
    if (request.method === "POST" && request.url === "/orders") {
      const raw = await bodyBytes(request);
      let input;
      try {
        input = JSON.parse(raw.toString());
      } catch {
        return reply(response, 400, { error: "invalid_json" });
      }
      const orderID = input?.order_id;
      if (typeof orderID !== "string" || !/^order_[a-zA-Z0-9_-]{1,40}$/.test(orderID) || orders.has(orderID)) {
        return reply(response, 400, { error: "invalid_or_duplicate_order" });
      }
      const session = await stripe.checkout.sessions.create({
        mode: "payment",
        client_reference_id: orderID,
        success_url: "http://localhost/success",
        cancel_url: "http://localhost/cancel",
      });
      const order = { order_id: orderID, session_id: session.id, status: "pending", paid_updates: 0, processed_events: 0 };
      orders.set(orderID, order);
      return reply(response, 201, order);
    }
    const match = /^\/orders\/(order_[a-zA-Z0-9_-]+)(\/session)?$/.exec(request.url);
    if (request.method === "GET" && match) {
      const order = orders.get(match[1]);
      if (!order) return reply(response, 404, { error: "order_not_found" });
      if (match[2]) {
        const session = await stripe.checkout.sessions.retrieve(order.session_id);
        return reply(response, 200, { id: session.id, client_reference_id: session.client_reference_id, payment_status: session.payment_status });
      }
      return reply(response, 200, order);
    }
    if (request.method === "POST" && request.url === "/webhooks/stripe") {
      const raw = await bodyBytes(request);
      let event;
      try {
        event = stripe.webhooks.constructEvent(raw, request.headers["stripe-signature"], webhookSecret);
      } catch {
        return reply(response, 400, { error: "invalid_signature" });
      }
      const session = event.data?.object;
      const order = orders.get(session?.client_reference_id);
      if (!order || order.session_id !== session.id || session.object !== "checkout.session") {
        return reply(response, 400, { error: "unknown_session" });
      }
      if (processedEvents.has(event.id)) return reply(response, 200, { duplicate: true });
      if (event.type === "checkout.session.completed" && session.payment_status === "paid") {
        if (order.status === "pending") {
          order.status = "paid";
          order.paid_updates++;
        }
      } else if (event.type === "checkout.session.async_payment_failed" && order.status === "pending") {
        order.status = "failed";
      }
      processedEvents.add(event.id);
      order.processed_events++;
      return reply(response, 200, { received: true });
    }
    return reply(response, 404, { error: "not_found" });
  } catch (error) {
    if (error instanceof RequestTooLargeError) return reply(response, 413, { error: "request_too_large" });
    return reply(response, 502, { error: "provider_or_application_error", detail: String(error.message) });
  }
});

server.listen(Number(process.env.PORT || 33001), process.env.HOST || "127.0.0.1");

import Stripe from "stripe";

const mockportURL = new URL(
  process.env.MOCKPORT_BASE_URL || "http://127.0.0.1:43101",
);
const stripe = new Stripe("sk_test_mockport", {
  apiVersion: "2025-10-29.clover",
  host: mockportURL.hostname,
  port: Number(mockportURL.port),
  protocol: mockportURL.protocol.replace(":", ""),
  telemetry: false,
});

const session = await stripe.checkout.sessions.create({
  mode: "payment",
  line_items: [{ price_data: { currency: "usd", unit_amount: 1200, product_data: { name: "Mockport item" } }, quantity: 1 }],
  client_reference_id: "example-cart",
  success_url: "http://localhost/success",
  cancel_url: "http://localhost/cancel",
});

console.log(JSON.stringify(session, null, 2));

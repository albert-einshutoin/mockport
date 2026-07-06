"use strict";

const fs = require("node:fs");
const path = require("node:path");
const Stripe = require("stripe");

function stripeSDKLabel() {
  return `stripe@${findPackageVersion(require.resolve("stripe"), "stripe")}`;
}

function findPackageVersion(entrypoint, packageName) {
  let dir = path.dirname(entrypoint);
  const root = path.parse(dir).root;

  while (dir !== root) {
    const packagePath = path.join(dir, "package.json");
    if (fs.existsSync(packagePath)) {
      const packageData = JSON.parse(fs.readFileSync(packagePath, "utf8"));
      // Stripe resolves through cjs/ first, so verify the package name before trusting the version field.
      if (packageData.name === packageName && packageData.version) {
        return packageData.version;
      }
    }
    dir = path.dirname(dir);
  }

  throw new Error(`could not locate ${packageName} package metadata`);
}

async function runStripeSmoke(options) {
  const base = new URL(options.baseURL);
  const stripe = new Stripe("sk_test_mockport", {
    apiVersion: "2025-10-29.clover",
    host: base.hostname,
    port: Number(base.port),
    protocol: base.protocol.replace(":", ""),
    telemetry: false,
  });

  const checkout = await stripe.checkout.sessions.create({
    mode: "payment",
    client_reference_id: "cart_sdk_1",
    success_url: "http://localhost/success",
    cancel_url: "http://localhost/cancel",
  });
  const retrievedCheckout = await stripe.checkout.sessions.retrieve(checkout.id);
  const checkoutList = await stripe.checkout.sessions.list({ limit: 10 });

  const paymentIntent = await stripe.paymentIntents.create({
    amount: 1200,
    currency: "usd",
  });
  const retrievedPaymentIntent = await stripe.paymentIntents.retrieve(paymentIntent.id);
  const paymentIntentList = await stripe.paymentIntents.list({ limit: 10 });

  const customer = await stripe.customers.create({ email: "customer@example.test" });
  const retrievedCustomer = await stripe.customers.retrieve(customer.id);
  const product = await stripe.products.create({ name: "Mockport Product" });
  const price = await stripe.prices.create({
    product: product.id,
    currency: "usd",
    unit_amount: 1200,
  });
  const subscription = await stripe.subscriptions.create({
    customer: customer.id,
    items: [{ price: price.id }],
  });
  const invoice = await stripe.invoices.create({ customer: customer.id });
  const refund = await stripe.refunds.create({ payment_intent: paymentIntent.id });

  assertEqual(retrievedCheckout.id, checkout.id, "retrieved checkout id");
  assertEqual(checkoutList.data[0].id, checkout.id, "checkout list first id");
  assertEqual(retrievedPaymentIntent.id, paymentIntent.id, "retrieved payment intent id");
  assertEqual(paymentIntentList.data[0].id, paymentIntent.id, "payment intent list first id");
  assertEqual(retrievedCustomer.id, customer.id, "retrieved customer id");
  assertEqual(price.product, product.id, "price product id");
  assertEqual(subscription.customer, customer.id, "subscription customer id");
  assertEqual(invoice.customer, customer.id, "invoice customer id");
  assertEqual(refund.payment_intent, paymentIntent.id, "refund payment intent id");

  // X-Mockport-Scenario ヘッダによる per-request シナリオ切り替えのテスト。
  // Stripe SDK は per-request カスタムヘッダをサポートしていないため fetch で直接確認する。
  const scenarioRes = await fetch(`${options.baseURL}/v1/checkout/sessions`, {
    method: "POST",
    headers: {
      "Authorization": "Bearer sk_test_mockport",
      "Content-Type": "application/x-www-form-urlencoded",
      "X-Mockport-Scenario": "payment_failed",
    },
    body: "mode=payment&success_url=http%3A%2F%2Flocalhost%2Fsuccess&cancel_url=http%3A%2F%2Flocalhost%2Fcancel",
  });
  const scenarioBody = await scenarioRes.json();
  assertEqual(scenarioRes.status, 402, "X-Mockport-Scenario: payment_failed returns 402");
  assertEqual(scenarioBody.error?.code, "card_declined", "payment_failed error code");

  return {
    provider: "stripe",
    baseURL: options.baseURL,
    status: "sdk-ok",
    sdk: stripeSDKLabel(),
    checkoutSession: checkout.id,
    paymentIntent: paymentIntent.id,
    customer: customer.id,
    product: product.id,
    price: price.id,
    subscription: subscription.id,
    invoice: invoice.id,
    refund: refund.id,
  };
}

function assertEqual(got, want, label) {
  if (got !== want) {
    throw new Error(`${label}: got ${got}, want ${want}`);
  }
}

module.exports = { runStripeSmoke };

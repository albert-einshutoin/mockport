"use strict";

const Stripe = require("stripe");

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

  return {
    provider: "stripe",
    baseURL: options.baseURL,
    status: "sdk-ok",
    sdk: "stripe@22.2.0",
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

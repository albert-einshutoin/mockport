# Node SDK Clients

These runnable examples connect the official Stripe, OpenAI, and LINE Node.js
SDKs to one local Mockport process. They never contact provider APIs and use only
fake credentials.

From the repository root, start Mockport:

```bash
go run ./cmd/mockport run --config examples/multi-adapter/mockport.yml
```

In another terminal:

```bash
cd examples/node-sdk-clients
npm ci
npm run stripe
npm run openai
npm run line
```

Set `MOCKPORT_BASE_URL` to use a non-default local address. Keep it on loopback;
these examples are intended for local development and CI only.

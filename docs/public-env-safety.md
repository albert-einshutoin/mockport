# Public Env Safety

[日本語版](public-env-safety.ja.md)

Mockport generated env files are intended to be safe to commit when they remain unchanged.

## Safe Values

Safe Mockport env values use:

- `mockport_` fake credential prefixes.
- `whsec_mockport` for fake webhook signing.
- `http://localhost:43101` or other local-only URLs.

Examples:

```env
STRIPE_API_URL=http://localhost:43101/stripe
STRIPE_SECRET_KEY=mockport_stripe_secret
STRIPE_WEBHOOK_SECRET=whsec_mockport
OPENAI_BASE_URL=http://localhost:43101/openai/v1
OPENAI_API_KEY=mockport_openai_key
GITHUB_OAUTH_BASE_URL=http://localhost:43101/github
GITHUB_OAUTH_CLIENT_ID=mockport_github_client
GITHUB_OAUTH_CLIENT_SECRET=mockport_github_secret
SLACK_BOT_TOKEN=mockport_slack_token
LINE_API_BASE_URL=http://localhost:43101/line
LINE_CHANNEL_TOKEN=mockport_line_channel_token
```

## Unsafe Values

Do not commit:

- Real provider API keys or tokens.
- Production provider API URLs.
- Customer payloads.
- Captured webhook secrets.
- Ambiguous placeholders that may later be replaced with real credentials.

## Check

```bash
bash scripts/check-public-env.sh
```

This check scans public docs, `.ja.md` translations, GitHub templates/workflows, example configs, packaging docs, and contract docs for real-looking provider credentials, production provider URLs, and ambiguous placeholders.

Detector-reference docs and intentionally unsafe warning fixtures must wrap unsafe examples in a narrow `mockport-public-safety` allow block:

```md
<!-- mockport-public-safety: allow-begin detector-reference -->
unsafe detector example only
<!-- mockport-public-safety: allow-end -->
```

Do not use allow blocks for normal setup instructions, generated examples, or user-facing quickstarts.

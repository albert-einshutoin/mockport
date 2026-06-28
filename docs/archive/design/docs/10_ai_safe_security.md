# 10. AI-safe Security

[日本語版](10_ai_safe_security.ja.md)

## Security principle

Mockport should not try to hide real secrets from AI. It should make real secrets unnecessary.

## AI-safe mode

Default:

```yaml
mode: ai-safe
```

## What AI-safe mode does

- warns on real-looking secrets
- warns on real external service URLs
- redacts all secret-like values in logs
- prevents proxying to known real services unless explicitly allowed
- generates fake local secrets
- marks reports as safe/unsafe

## Dangerous secret patterns

Initial examples:

<!-- mockport-public-safety: allow-begin detector-reference -->
```txt
sk_live_
sk_test_
AKIA
ASIA
ghp_
github_pat_
xoxb-
xoxp-
AIza
whsec_
```
<!-- mockport-public-safety: allow-end -->

Important nuance:

Some providers use `test` secrets that are still real credentials. Mockport should warn on both live and test provider keys.

## Fake secret prefixes

Recommended fake prefixes:

```txt
mockport_
local_
fake_
dummy_
```

Examples:

```env
STRIPE_SECRET_KEY=mockport_stripe_secret
OPENAI_API_KEY=mockport_openai_key
LINE_CHANNEL_SECRET=mockport_line_secret
```

## Dangerous URL patterns

Examples:

<!-- mockport-public-safety: allow-begin detector-reference -->
```txt
https://api.stripe.com
https://api.openai.com
https://api.github.com
https://api.line.me
https://slack.com/api
```
<!-- mockport-public-safety: allow-end -->

## Strict mode

Strict mode should fail startup:

```yaml
mode: strict
```

If real-looking secrets or real service URLs are detected, Mockport exits with non-zero status.

## Redaction

Never output full secret values.

Bad:

<!-- mockport-public-safety: allow-begin detector-reference -->
```txt
STRIPE_SECRET_KEY=sk_live_xxxxx
```
<!-- mockport-public-safety: allow-end -->

Good:

```txt
STRIPE_SECRET_KEY looks like a real Stripe key.
```

## Report fields

Report should include:

```txt
Safety:
- real-looking secrets detected
- external URLs detected
- fake secrets generated
- mode
```

## Non-goals

Mockport is not:

- a secret manager
- a DLP system
- a production security gateway
- a sandbox escape prevention system

It is a safer default for integration testing.

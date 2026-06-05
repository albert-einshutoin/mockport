# Mockport SDK Contract Harness

[日本語版](README.ja.md)

This workspace is test-only. It keeps future official SDK contract tests outside the Go runtime and outside the experimental npm wrapper.

## Commands

```bash
npm test
```

Runs the offline harness sanity check.

```bash
MOCKPORT_BASE_URL=http://127.0.0.1:43101 npm run test:live -- --provider all --json
```

Runs the live SDK contract against a local Mockport server. With `--provider all`,
the runner executes the real `stripe`, `openai`, `github-oauth`, and `slack` smoke
tests in order and aggregates their results under a `providers` array. If any
provider fails, the runner reports `status: "failed"` and exits non-zero.

When `--offline` is passed, the runner falls back to the placeholder sanity check
instead of contacting a server.

Supported provider selectors:

- `all` (runs every real smoke test below)
- `stripe`
- `openai`
- `github-oauth`
- `slack`

LINE is not yet supported: it has no smoke test in this workspace and is therefore
excluded from the `all` selector.

No test in this workspace may call an external provider API.

# Scenario Policy

[日本語版](scenario-policy.ja.md)

Mockport supports built-in scenarios first. User-defined scenarios are future work and must stay separate from provider compatibility claims until they are promoted through the fixture and compatibility process.

Built-in scenarios are **stable local contracts** for `workflow-compatible` testing. They define deterministic behavior inside Mockport, not full provider parity. User YAML scenarios may layer additional local behavior for app-specific tests, but they must not imply provider compatibility or hide documented gaps.

## Built-in Scenarios

Built-in scenarios are maintained by Mockport. Each one is published as a deterministic contract in the adapter spec (`docs/adapters/<adapter>.md`). They are the source of truth for what a built-in adapter can reproduce locally. They must have:

- A stable scenario name.
- Adapter metadata.
- Tests.
- Public-safe examples.
- Fixture or documentation evidence when used for compatibility scoring.

Built-in scenarios can contribute to compatibility scores only when they are backed by source metadata and visible in reports.

## User-defined Scenarios

User-defined scenarios are local project behavior. They are a **local test convenience** only and are **not a provider compatibility claim**. They may layer extra local responses or state for app-specific tests, but they do not prove provider compatibility by themselves and must not be documented as if they were built-in adapter contracts.

### Built-in vs user-defined

| Aspect | Built-in adapter scenarios | User-defined scenarios |
| --- | --- | --- |
| Source | Maintained by Mockport in adapter specs and tests | Defined locally in a project (future `scenarios:` block) |
| Contract | Deterministic contract published in `docs/adapters/<adapter>.md` | Project-local behavior; no public adapter contract |
| Compatibility impact | Can contribute to compatibility scoring when backed by tests, docs, and fixture evidence | Does not raise provider compatibility score or maturity |
| Runtime status | Active via `adapters.<name>.scenario` and `X-Mockport-Scenario` | `scenarios:` block is parsed but not implemented at runtime |

### Minimal config example

Built-in scenario selection uses the adapter's `scenario:` field. Examples in docs, fixtures, and config must use **fake local values only** (mock secrets, local URLs, synthetic IDs). Do not copy production credentials, customer payloads, or real provider endpoints into scenario examples.

```yaml
adapters:
  stripe:
    enabled: true
    base_path: /stripe
    scenario: payment_success
    fake_secret: mockport_stripe_secret
    webhook:
      target_url: http://app:3000/webhooks/stripe
      signing_secret: whsec_mockport
```

This selects the built-in `payment_success` contract from the Stripe adapter spec. It does not define a user-defined scenario.

Until a full user-defined scenario system exists, adapters should prefer explicit built-in scenarios over partial custom behavior. If a user-defined scenario is later promoted, it must receive a built-in scenario name, tests, docs, and sanitized fixture evidence.

> **Current status:** The `scenarios:` block in `mockport.yml` is parsed but not yet implemented at runtime. Mockport emits a warning when this block is present. See [limitations](site/limitations.md#unimplemented-configuration-blocks) for details.

## Compatibility Boundary

Compatibility scoring must distinguish:

- Mockport built-in scenario coverage.
- SDK contract coverage.
- Workflow state coverage.
- User-defined local behavior.

Only the first three can raise provider compatibility maturity. User-defined local behavior can be reported, but it must not hide unsupported provider behavior.

Unsupported provider behavior must stay documented as **known gaps** in adapter specs, the support matrix, and compatibility reports. Scenario docs and examples must not overclaim coverage or imply that a user-defined override closes a documented gap.

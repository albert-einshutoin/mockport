# Limitations

[日本語版](limitations.ja.md)

Mockport targets provider-compatible local APIs for selected workflows. It does not reproduce provider internals or undocumented behavior.

## Current Preview Scope

Current mainline adapters are workflow-compatible for selected local and CI integration paths. They are not a substitute for provider sandboxes or production validation.

## What Mockport Does Not Reproduce

- Real payment processing, fraud systems, settlement, or billing networks.
- Real AI inference, model quality, provider tokenization, or private scheduling behavior.
- Real GitHub organization, enterprise, or permission policy.
- Real Slack workspace delivery, enterprise policy, or full directory state.
- Real LINE Login UI, LIFF browser runtime, provider webhook redelivery, quota/rate-bucket enforcement, regional policy, or Dapp Portal behavior.
- Undocumented provider behavior.

## How To Evaluate Support

Use:

- [Support matrix](support-matrix.md)
- `/_mockport/report`
- `mockport report --format json`
- Adapter examples
- Public env safety checks

Unsupported endpoints and approximations should be visible before adoption.

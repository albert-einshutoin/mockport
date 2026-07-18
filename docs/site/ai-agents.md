# AI Coding Agents

[日本語版](ai-agents.ja.md)

Mockport lets coding agents exercise integration code without receiving real
provider credentials. Real keys should never be placed in prompts, generated
configuration, issue bodies, logs, or fixtures. Mockport's generated
`mockport_` values are deterministic fakes intended for local development and CI.

## Add project context

Generate the local configuration and print an instruction block for the adapters
your project uses:

```bash
mockport init --adapter stripe --adapter openai
mockport agent-context --adapter stripe --adapter openai >> AGENTS.md
```

The command writes only to stdout. It does not create or overwrite agent files,
so you can review the output before appending it to `AGENTS.md`, `CLAUDE.md`, or
another tool-specific instruction file.

## Example request to an agent

> Implement Stripe checkout against the local Mockport endpoint described in
> AGENTS.md. Use fake credentials only, test the payment failure path with
> X-Mockport-Scenario, and do not contact Stripe.

## Verification loop

1. Start Mockport with the generated Compose file.
2. Run the application or integration test written by the agent.
3. Read `GET http://localhost:43101/_mockport/report`.
4. Check `unsupported_endpoints`, `safety_warnings`, and request status codes.
5. Treat an unsupported endpoint as a scope mismatch; do not invent a successful response or add a real provider fallback.

Use [Support matrix](support-matrix.md) for current behavior and
[Limitations](limitations.md) for explicit non-goals. Planning documents under
`docs/planning/` are not current support commitments.

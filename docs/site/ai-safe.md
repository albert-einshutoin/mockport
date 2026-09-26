# AI-safe Mode

[日本語版](ai-safe.ja.md)

Mockport defaults to `ai-safe` mode. It warns on real-looking secrets and real external service URLs.

Use strict mode when startup should fail on unsafe config:

```yaml
mode: strict
```

Strict mode checks whether configuration contains unsafe credentials or external URLs. It does not verify incoming API keys. For Stripe and OpenAI, set `auth_required: true` alongside a `fake_secret` to reject missing or incorrect bearer keys on ordinary API requests. The runtime report shows `auth_required=false` when key verification was disabled.
Setting `auth_required: true` for another adapter is a configuration error.

Check config without starting a server:

```bash
mockport run --config mockport.yml --check
```

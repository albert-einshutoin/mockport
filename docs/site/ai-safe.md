# AI-safe Mode

[日本語版](ai-safe.ja.md)

Mockport defaults to `ai-safe` mode. It warns on real-looking secrets and real external service URLs.

Use strict mode when startup should fail on unsafe config:

```yaml
mode: strict
```

Check config without starting a server:

```bash
mockport run --config mockport.yml --check
```

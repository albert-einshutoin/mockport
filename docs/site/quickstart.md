# Quickstart

[日本語版](quickstart.ja.md)

```bash
mockport init --adapter stripe
docker compose -f docker-compose.mockport.yml up
curl http://localhost:43101/health
```

For multiple adapters:

```bash
mockport init --adapter stripe --adapter openai --adapter github-oauth --adapter slack --adapter line
docker compose -f docker-compose.mockport.yml up
```

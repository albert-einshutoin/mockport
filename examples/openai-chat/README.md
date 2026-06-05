# OpenAI Chat Example

[日本語版](README.ja.md)

This example runs the OpenAI-compatible Mockport adapter with fake local credentials.

```bash
docker build -t mockport:local -f docker/Dockerfile .
mockport run --config examples/openai-chat/mockport.yml
```

Use these values in the application under test:

```env
OPENAI_BASE_URL=http://localhost:43101/openai/v1
OPENAI_API_KEY=mockport_openai_key
```

Smoke test:

```bash
curl http://localhost:43101/openai/v1/models
curl -X POST http://localhost:43101/openai/v1/chat/completions
curl http://localhost:43101/_mockport/report
```

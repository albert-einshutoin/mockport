# GitHub OAuth Example

[日本語版](README.ja.md)

This example runs the GitHub OAuth-like Mockport adapter with fake local credentials.

```bash
docker build -t mockport:local -f docker/Dockerfile .
mockport run --config examples/github-oauth/mockport.yml
```

Use these values in the application under test:

```env
GITHUB_OAUTH_BASE_URL=http://localhost:43101/github
GITHUB_OAUTH_CLIENT_ID=mockport_github_client
GITHUB_OAUTH_CLIENT_SECRET=mockport_github_secret
```

Smoke test:

```bash
curl -i "http://localhost:43101/github/login/oauth/authorize?client_id=mockport_github_client&redirect_uri=http://localhost:3000/callback&state=local"
curl -X POST http://localhost:43101/github/login/oauth/access_token
curl http://localhost:43101/github/user
curl http://localhost:43101/_mockport/report
```

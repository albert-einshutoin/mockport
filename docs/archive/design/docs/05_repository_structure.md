> **⚠️ Archive notice — Not maintained, may diverge, do not cite as authoritative.**
>
> Pre-implementation design archive. This is **not** the authoritative source for current implementation.
> For current specs see [docs/site/](../../../site/index.md).

# 05. Repository Structure

[日本語版](05_repository_structure.ja.md)

Recommended initial single repository:

```txt
mockport/
  README.md
  README.ja.md
  LICENSE
  go.mod
  go.sum
  Makefile
  .gitignore
  .dockerignore

  cmd/
    mockport/
      main.go

  internal/
    app/
      app.go

    cli/
      root.go
      init.go
      run.go
      detect.go
      add.go
      report.go

    config/
      config.go
      loader.go
      validate.go
      defaults.go

    server/
      server.go
      router.go
      middleware.go
      health.go

    adapter/
      adapter.go
      registry.go

    scenario/
      scenario.go
      store.go
      response.go

    report/
      recorder.go
      report.go
      redactor.go

    detector/
      detector.go
      env.go
      package_json.go
      compose.go

    docker/
      compose.go
      templates.go

    security/
      secrets.go
      urls.go
      aisafe.go

  adapters/
    stripe/
      adapter.go
      routes.go
      models.go
      scenarios.go
      webhook.go
      signatures.go
      adapter_test.go

  configs/
    mockport.example.yml

  examples/
    stripe-checkout/
      README.md
      docker-compose.yml
      mockport.yml
      .env.mockport.example

  docker/
    Dockerfile

  docs/
    architecture.md
    quickstart.md
    adapter-development.md
    ai-safe-development.md

  .github/
    workflows/
      ci.yml
      docker.yml
```

## Why single repository first

Mockport should start as a single repository because:

- core and adapters will evolve together
- breaking interface changes are likely early
- CI is simpler
- documentation is easier to keep consistent
- the all-in-one Docker image is simpler for early users

## When to split

Split only after:

- adapter interface stabilizes
- multiple external contributors exist
- Docker image size becomes a real problem
- independent adapter release cycles become necessary

## Future split model

```txt
github.com/mockport/mockport
github.com/mockport/adapter-stripe
github.com/mockport/adapter-openai
github.com/mockport/adapter-github
github.com/mockport/docs
```

Do not start here.

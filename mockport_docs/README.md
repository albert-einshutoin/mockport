# Mockport Development Documentation Pack

[日本語版](README.ja.md)

Created: 2026-05-25

Mockport is a Docker-first service emulator for secret-free integration testing in AI-native development.

This documentation pack defines the initial product concept, MVP scope, Go implementation strategy, architecture, adapter model, CLI commands, Docker usage, security philosophy, and execution roadmap.

## Recommended stack

- Go: 1.26.3
- Docker Engine: 29.5.2
- Docker Compose: v5.x line
- Runtime distribution: Docker image first, Go binary second, optional npm wrapper later
- Repository model: single repository / monorepo for MVP

## Documents

1. `docs/00_product_vision.md`
2. `docs/01_problem_statement.md`
3. `docs/02_mvp_scope.md`
4. `docs/03_architecture.md`
5. `docs/04_go_engineering_guide.md`
6. `docs/05_repository_structure.md`
7. `docs/06_config_spec.md`
8. `docs/07_adapter_design.md`
9. `docs/08_cli_spec.md`
10. `docs/09_docker_runtime.md`
11. `docs/10_ai_safe_security.md`
12. `docs/11_testing_strategy.md`
13. `docs/12_roadmap.md`
14. `docs/13_readme_draft.md`
15. `docs/14_open_questions.md`
16. `docs/99_sources.md`

## First implementation target

The first implementation target is a Stripe-like adapter that can run in Docker and return payment success/failure responses without real Stripe secrets.

The goal is not to clone Stripe completely. The goal is to make common integration paths testable with fake local secrets and scenario-driven responses.

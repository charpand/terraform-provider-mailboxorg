# AI Agent Instructions

Applies to any AI agent working in this repository.

## Layout

- `internal/client/` — Terraform-independent mailbox.org JSON-RPC client; area packages (`domains`, …) wrap single API methods.
- `internal/provider/` — Terraform provider, resources and data sources.
- `internal/testutils/` — fake JSON-RPC server for tests.

## Rules

- API functions take `ctx context.Context, c *client.Client` first.
- Use `c.Read` for repeatable calls and `c.Mutate` for calls with side effects. Never replay a mutation automatically.
- Never put credentials, sessions or request parameters in errors or logs.
- Do not guess undocumented API behaviour (error codes, field types); verify it and note the evidence in the commit or PR.
- Tests must not need credentials or network; use `testutils.NewFakeServer`.
- Conventional commits, one concern each. No AI attribution lines.

## Workflow

`./scripts/format`, `./scripts/lint`, `./scripts/test`, and `./scripts/docs` after schema changes (CI fails on docs drift). Update `CHANGELOG.md`.

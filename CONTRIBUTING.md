# Contributing to Clyra

Clyra is an open-source project maintained in spare time. Contributions
are welcome. Keeping the maintainer burden low is important — a few
norms:

## Before you open a PR

- **File an issue first for anything non-trivial.** New features, API
  changes, or refactors should start as a discussion. A one-line typo
  fix doesn't need a prior issue.
- **One PR, one concern.** Mix a bug fix and a feature in the same PR
  and I'll ask you to split it.
- **Branch from `main`.** The repository has only one long-lived branch.

## What tests + checks must pass

- `go test ./...` — all backend packages, including the server, must stay
  green.
- `cd web && npm test` — frontend tests via Vitest.
- `cd web && npm run build` — the production bundle must compile.
- `cd web && npm run lint` — ESLint with the repo's flat config.
- `gofmt -s -w .` on anything you touched on the backend.
- `cd web && npm run build:catalog` **and commit the generated files** if
  you edited `components/*.md`.

CI (`.github/workflows/ci.yml`) runs these on every push and PR. A PR
that doesn't go green locally shouldn't be opened.

## Scope that's in-bounds

Any of these:
- Bug fixes in the editor, backend, or IaC renderers.
- New components in `components/*.md` with their aliases.
- New cost catalog entries (edit `web/src/lib/costCatalog.ts` and update
  `LAST_REVIEWED`).
- New architecture rules in `web/src/lib/architectureRules.ts`.
- Self-hosting documentation improvements.
- Additional MCP tools (follow the pattern in `internal/mcp/tools.go`;
  anything that calls managed AI needs the `ChargesManagedAI` flag).
- New LLM provider adapters in `internal/provider/`.
- Reverse-ingestion coverage for more Terraform resources.

## Scope that's out-of-bounds

- Changes that re-enable or hard-wire payment providers in a way that
  can't be turned off. The code is here if you want to self-host with
  paid tiers; it should stay opt-in via config, not required.
- Breaking API changes to the editor's cloud-sync format without a
  migration path for existing user_projects rows.

## Design-level changes

Open an issue tagged `rfc` with:
- The problem you're solving.
- Two or three concrete approaches with trade-offs.
- Your recommendation and why.

A short "yes let's do this" response in the issue is enough to proceed.

## Code style

- **Go**: follow stdlib idioms. Short names in small scopes. Return
  errors; don't swallow them. Avoid interface pollution — concrete types
  until a second caller shows up.
- **TypeScript**: strict mode. No `any` without a comment explaining
  why. Prefer type aliases over interfaces for data shapes.
- **Comments**: explain *why*, not *what*. A function named
  `extractText` doesn't need "// extracts text". It does need "// strips
  the ```json fence that some providers wrap single-object responses in"
  if that's load-bearing context.

## License

By contributing you agree your contribution is licensed under the
project's Apache 2.0 license. No separate CLA.

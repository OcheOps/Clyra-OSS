# Changelog

All notable changes to Clyra are tracked here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this
project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

Nothing yet.

## [0.1.0] — 2026-10-24

First open-source release. Clyra ran as a hosted SaaS at useclyra.com
from April through October 2026; this release is the continuation of
that codebase as a self-hostable free project.

### Added

- Full backend (Go) + SPA (React + Vite) codebase.
- Prompt → diagram pipeline across OpenAI, Anthropic, and Gemini.
- Reverse ingestion: paste Terraform, see the architecture graph.
- 1,000+ component catalog across AWS, GCP, Azure, Kubernetes,
  Cloudflare, with locally vendored AWS + GCP SVG icons.
- Live cost estimation for the most common SKUs.
- Well-Architected validator with 10 rule families.
- Blast radius analysis — "what breaks if this fails."
- MCP server at `POST /mcp` using the official Go SDK. One tool
  registered in OSS builds: `reverse_from_iac`.
- Session auth, OAuth (Google, GitHub), Personal Access Tokens.
- Cloud project storage backed by Postgres + optional S3-compatible
  object store.
- Append-only database migration ledger with advisory lock.
- Secret-pattern scrubber on the slog → Sentry handler
  (OpenAI, Anthropic, Google, AWS, PATs, Polar, Bearer).
- Rate limiting with per-IP, per-user, and per-user-per-day scopes.
- CI pipeline (Go tests, frontend lint + tests + build, generated
  artifact drift check).

### Disabled by default

- Managed Bedrock provider — the hosted managed-inference path is
  retired. The code remains in `internal/provider/bedrock.go` and the
  registration block in `cmd/server/main.go`; uncomment to re-enable
  with your own AWS credentials.
- Paystack and Polar payment providers — code preserved for reference.
  Uncomment and set the required env vars to run a paid tier.
- MCP `generate_diagram` tool — depends on managed AI. The local
  `reverse_from_iac` tool stays live.
- Public landing-demo endpoint — now short-circuits to a canned
  example diagram.

[Unreleased]: https://github.com/OcheOps/Clyra-OSS/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/OcheOps/Clyra-OSS/releases/tag/v0.1.0

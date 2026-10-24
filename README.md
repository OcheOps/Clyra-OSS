# Clyra

[![CI](https://github.com/OcheOps/Clyra-OSS/actions/workflows/ci.yml/badge.svg)](https://github.com/OcheOps/Clyra-OSS/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.25-00ADD8?logo=go&logoColor=white)](go.mod)
[![React](https://img.shields.io/badge/react-19-61DAFB?logo=react&logoColor=white)](web/package.json)
[![PRs welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

**Prompt → editable cloud architecture diagram → deployable Terraform.**

> **Watch the 40-second overview** — see
> [v0.1.0 release](https://github.com/OcheOps/Clyra-OSS/releases/tag/v0.1.0)
> for `brag-v2.mp4`.

Clyra turns a one-line description of a system into an editable
multi-cloud architecture diagram with cost estimates, Well-Architected
validation, and one-click Terraform / CloudFormation export. It also goes
the other way — paste existing Terraform and get the graph.

Originally shipped as a hosted SaaS at useclyra.com. **The hosted product
has been retired as of October 2026**; the codebase is now open source
under Apache 2.0.

## Status

This is a working codebase, not a toy. What's in the box:

- **Go backend** with session auth, OAuth (Google, GitHub), personal
  access tokens, cloud project storage (Postgres + S3-compatible), rate
  limiting (Redis), and audit logging.
- **React + Vite SPA** with a React Flow editor, 1,000+ component catalog
  (AWS, GCP, Azure, Kubernetes, Cloudflare), live cost panel, blast
  radius analysis, and vendored AWS/GCP icons.
- **AI providers**: OpenAI, Anthropic, Gemini — bring your own key.
  Managed Bedrock exists in the code and is deliberately disabled
  (`cmd/server/main.go`) — a self-hoster who wants a managed path can
  re-enable it with their own AWS credentials.
- **MCP server** at `POST /mcp` using the official Go SDK. Ships one tool
  (`reverse_from_iac`); `generate_diagram` is disabled in OSS builds
  since it depended on managed inference.
- **Reverse IaC → diagram** parsing. Paste a Terraform file, see the
  architecture.

What's not in the box: a hosted managed-AI tier, a payment flow (the
Paystack and Polar wiring is commented out but preserved for reference),
or a public demo API.

## Quick start (dev)

```bash
# Backend
cp .env.example .env    # fill in POSTGRES_* and at least one of OPENAI/ANTHROPIC/GEMINI
docker compose up -d postgres redis
go run ./cmd/server     # http://localhost:8080

# Frontend
cd web
npm install
npm run build:catalog   # regenerate component catalog
npm run dev             # http://localhost:3000
```

Open `http://localhost:3000`, create an account, drop your OpenAI /
Anthropic / Gemini key into **Settings → BYO keys**, and start generating.

## Self-hosting

Docker Compose provides Postgres and Redis for local dev. For production,
build the server + SPA with `docker build .` (ships both in one image via
multi-stage build) and point your reverse proxy at port 8080.

Required env vars (minimum):
- `POSTGRES_URL` — connection string
- `REDIS_URL` — for rate limiting + MCP daily quotas
- `SESSION_SECRET` — 32+ random bytes
- `CORS_ALLOWED_ORIGINS` — e.g. `https://your-clyra.example.com`

Optional:
- `OAUTH_GOOGLE_*`, `OAUTH_GITHUB_*` — social login
- `RESEND_API_KEY` — transactional emails (invites, password reset)
- `SENTRY_DSN` — error reporting (with built-in secret scrubbing)
- `S3_*` — if using an S3-compatible object store for cloud project storage

See `.env.example` for the full list.

## Repository layout

```
cmd/
  server/            HTTP server entry point
  migrate/           Standalone migration runner
internal/
  auth/              Session + OAuth + PAT
  diagram/           Prompt → diagram pipeline (service, prompt, normalizer)
  iac/               Diagram → Terraform / CloudFormation render + reverse parse
  mcp/               MCP server, tools, rate limiter
  provider/          LLM provider adapters (openai, anthropic, gemini, bedrock)
  payments/          Paystack + Polar wiring (code preserved, not wired)
  publicdemo/        Landing demo endpoint (short-circuits to canned diagram)
  observability/     Slog → Sentry handler with secret-pattern scrubber
web/
  src/features/      Editor, settings, admin, auth, billing, etc.
  src/lib/           Generated component catalog, cost catalog, icons
components/          Markdown source files for the catalog generator
scripts/             Catalog generation, icon vendoring, cost coverage
docs/                MCP spec
brag-v2/             Remotion source for the project overview video
```

## MCP

Clyra ships an MCP server at `POST /<your-host>/mcp` using the official
Go SDK's Streamable HTTP transport. Mint a PAT at `/settings/mcp` and
drop this into Claude Desktop / Cursor / any MCP-capable client:

```json
{
  "mcpServers": {
    "clyra": {
      "url": "https://your-clyra-instance.example.com/mcp",
      "headers": { "Authorization": "Bearer clyra_pat_…" }
    }
  }
}
```

Tools registered in OSS builds: `reverse_from_iac` (paste HCL → get the
graph, no AI involved). The `generate_diagram` tool exists but isn't
registered — see `cmd/server/main.go` to turn it on if you've wired your
own provider.

See [docs/mcp.md](docs/mcp.md) for the full protocol notes.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Short version: file an issue
before non-trivial PRs; `go test ./...` and `npm test` must pass; use
`gofmt` and ESLint defaults.

## Security

Report vulnerabilities privately via GitHub Security Advisories, not in
public issues. See [SECURITY.md](SECURITY.md).

## License

[Apache License 2.0](LICENSE).

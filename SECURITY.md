# Security policy

## Reporting a vulnerability

Open a **private security advisory** on GitHub rather than filing a
public issue:

<https://github.com/OcheOps/Clyra-OSS/security/advisories/new>

Include:
- Which component is affected (backend package, frontend file, Docker
  image, or docs).
- A concrete reproducer.
- Impact — read, modification, denial of service, auth bypass, exfil.
- Any suggested fix.

A triage reply usually comes within a few days. Fix timelines depend on
severity.

## Threat model

Clyra is a self-hosted application. Each deployment is one operator's
responsibility. The code assumes:

- **The server operator is trusted.** Server-side logs may contain
  user prompts, generated diagrams, email addresses, and audit events.
  Keep the Postgres database, Redis, and env vars locked down.
- **Users are not trusted with raw provider keys.** The BYO flow sends
  the user's OpenAI / Anthropic / Gemini key to the backend over
  TLS; the backend forwards it to the provider for the single request
  and never persists it. On the client, BYO keys live in `localStorage`
  scoped per account (`clyra-byo-keys:<userId>`).
- **MCP callers are plan-gated.** The /mcp endpoint requires a Personal
  Access Token (`clyra_pat_…`, hashed server-side) and respects the same
  plan gate as in-app generation. See `docs/mcp.md`.

## Defenses in place

- **CSRF**: double-submit cookie on every mutating request. GETs skip it.
- **CORS**: strict allow-list (`CORS_ALLOWED_ORIGINS`); `*` rejected in
  production.
- **CSP**: `script-src 'self'` (no unsafe-eval, no unsafe-inline);
  `connect-src` scoped to the hosts the SPA actually calls.
- **HSTS**: 6 months in production.
- **Rate limiting**: per-IP for auth routes, per-user for generation,
  per-user daily for MCP managed-AI tools. Redis-backed.
- **Secret scrubbing**: `internal/observability/scrub.go` redacts known
  key shapes (OpenAI, Anthropic, Google, AWS, PATs, Polar webhooks,
  Bearer headers) in every slog Error record before it reaches Sentry.
- **Audit log**: billing, plan changes, admin actions, and MCP tool
  calls land in `audit_events`.
- **Gemini keys** go in the `x-goog-api-key` header, not the URL query
  string, so a network error from Google doesn't stringify the key into
  our error path.

## Known operational caveats

- BYO keys in `localStorage` are unencrypted — standard browser-side
  risk (malware, extensions, physical access).
- The MCP memory-only rate limiter resets on process restart; the
  Redis-backed limiter (default when `REDIS_URL` is set) is correct
  across restarts and replicas.
- The managed-Bedrock provider is disabled by default in OSS builds. If
  a self-hoster re-enables it with their own AWS credentials, they take
  on the operational responsibility of that spend.

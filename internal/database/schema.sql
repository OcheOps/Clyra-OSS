CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'user',
    plan          TEXT NOT NULL DEFAULT 'free',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

CREATE TABLE IF NOT EXISTS sessions (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);

CREATE TABLE IF NOT EXISTS usage_records (
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    usage_date    DATE NOT NULL DEFAULT CURRENT_DATE,
    request_count INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, usage_date)
);

CREATE INDEX IF NOT EXISTS idx_usage_records_date ON usage_records (usage_date);

CREATE TABLE IF NOT EXISTS usage_log (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider   TEXT NOT NULL,
    success    BOOLEAN NOT NULL,
    latency_ms INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_usage_log_user_date ON usage_log (user_id, created_at);

CREATE TABLE IF NOT EXISTS invite_codes (
    id           TEXT PRIMARY KEY,
    code         TEXT NOT NULL UNIQUE,
    created_by   TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    used_by      TEXT REFERENCES users(id) ON DELETE SET NULL,
    used_at      TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    plan         TEXT NOT NULL DEFAULT 'free',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invite_codes_code ON invite_codes (code);

CREATE TABLE IF NOT EXISTS invite_requests (
    id         TEXT PRIMARY KEY,
    email      TEXT NOT NULL,
    note       TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL DEFAULT 'pending', -- pending | approved | declined
    handled_by TEXT REFERENCES users(id) ON DELETE SET NULL,
    handled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invite_requests_status ON invite_requests (status, created_at DESC);

-- Payments (one-time plan upgrades for now; no subscriptions).
CREATE TABLE IF NOT EXISTS payments (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider     TEXT NOT NULL,                      -- 'paystack' | 'polar' | 'mock'
    provider_ref TEXT NOT NULL UNIQUE,               -- our internal reference (clyra-xxx)
    external_id  TEXT,                               -- provider-native ID (Polar subscription UUID; NULL for Paystack)
    amount_minor INT NOT NULL,                       -- smallest currency unit (kobo / cents)
    currency     TEXT NOT NULL,                      -- 'NGN' | 'USD'
    region       TEXT NOT NULL,                      -- 'NG' | 'US' | ''
    plan         TEXT NOT NULL,                      -- target plan on success
    period_days  INT NOT NULL DEFAULT 30,
    status       TEXT NOT NULL DEFAULT 'pending',    -- 'pending' | 'paid' | 'failed' | 'refunded'
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at  TIMESTAMPTZ
);

-- Additive migration for existing DBs that predate external_id.
ALTER TABLE payments ADD COLUMN IF NOT EXISTS external_id TEXT;

-- Polar order_id captured on verify. Required for Polar's /v1/refunds/ API
-- which keys on order_id specifically (not subscription_id, not checkout_id).
-- NULL for Paystack — its ProviderRef already works for refunds.
ALTER TABLE payments ADD COLUMN IF NOT EXISTS order_id TEXT;

-- Paystack subscription email_token — required alongside subscription_code
-- when calling /subscription/disable. Comes back on the charge.success
-- (initial) and subscription.create webhooks. NULL for Polar and for
-- Paystack one-time charges made before subscriptions were introduced.
ALTER TABLE payments ADD COLUMN IF NOT EXISTS email_token TEXT;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS checkout_id TEXT;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS recurring BOOLEAN NOT NULL DEFAULT FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_checkout_id ON payments (checkout_id) WHERE checkout_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_order_id ON payments (order_id) WHERE order_id IS NOT NULL;

-- Personal Access Tokens for the MCP server (docs/mcp.md). Users mint one
-- from /settings/mcp and paste into Claude Desktop's config. We store only
-- sha256(clear_token); the cleartext is shown once at generation time and
-- never again.
--
-- source distinguishes user-minted PATs from OAuth-issued tokens (phase 4).
-- last_used is stamped by the auth middleware so the UI can show freshness
-- and users can spot stale entries to revoke.
CREATE TABLE IF NOT EXISTS api_tokens (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,                    -- user-supplied label
    token_hash TEXT NOT NULL UNIQUE,             -- sha256(clear_token) hex
    source     TEXT NOT NULL DEFAULT 'pat',      -- 'pat' | 'oauth'
    last_used  TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

-- Partial index over live tokens only — lookups filter revoked_at IS NULL,
-- and revoked rows aren't hot data (kept for audit / reactivation checks).
CREATE INDEX IF NOT EXISTS idx_api_tokens_hash_live ON api_tokens (token_hash)
    WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_tokens_user ON api_tokens (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payments_user ON payments (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments (status);

-- Cloud-synced projects. Metadata lives here; the diagram payload may live in S3
-- (when configured) or inline in the `data` column (dev fallback).
CREATE TABLE IF NOT EXISTS user_projects (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title         TEXT NOT NULL DEFAULT 'Untitled Diagram',
    version       INT  NOT NULL DEFAULT 1,
    storage       TEXT NOT NULL DEFAULT 'db',   -- 'db' | 's3'
    s3_key        TEXT,
    preview_key   TEXT,
    data          JSONB,                         -- populated when storage='db'
    byte_size     INT  NOT NULL DEFAULT 0,
    preview_bytes INT  NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_projects_user ON user_projects (user_id, updated_at DESC);

-- Additive migrations for the projects table (idempotent).
ALTER TABLE user_projects ADD COLUMN IF NOT EXISTS title         TEXT NOT NULL DEFAULT 'Untitled Diagram';
ALTER TABLE user_projects ADD COLUMN IF NOT EXISTS preview_key   TEXT;
ALTER TABLE user_projects ADD COLUMN IF NOT EXISTS preview_bytes INT  NOT NULL DEFAULT 0;

-- MCP: tag rows created via the MCP server so the UI can badge them
-- ('Created via MCP') and analytics can measure the outside-in wedge.
-- 'app' covers everything created through the SPA (default for legacy rows).
ALTER TABLE user_projects ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'app';
ALTER TABLE user_projects DROP CONSTRAINT IF EXISTS user_projects_source_check;
ALTER TABLE user_projects ADD CONSTRAINT user_projects_source_check
    CHECK (source IN ('app', 'mcp'));

-- Password reset tokens. The plaintext token is emailed to the user; we
-- only persist its SHA-256 hash so a DB leak can't impersonate anyone.
CREATE TABLE IF NOT EXISTS password_resets (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_resets_user ON password_resets (user_id, created_at DESC);

-- Additive migrations for older DBs (idempotent).
ALTER TABLE users        ADD COLUMN IF NOT EXISTS plan TEXT NOT NULL DEFAULT 'free';
ALTER TABLE users        ADD COLUMN IF NOT EXISTS plan_expiry TIMESTAMPTZ;
ALTER TABLE invite_codes ADD COLUMN IF NOT EXISTS plan TEXT NOT NULL DEFAULT 'free';

-- OAuth identities. A user can have zero (password-only), one, or many rows
-- here — same verified email auto-links a new provider to the existing user.
-- password_hash becomes nullable because OAuth-only users never set one.
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

CREATE TABLE IF NOT EXISTS user_identities (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider     TEXT NOT NULL,                   -- 'google' | 'github'
    provider_sub TEXT NOT NULL,                   -- provider's stable user id
    email        TEXT NOT NULL,                   -- email at link time (for audit; not authoritative)
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_sub)
);

CREATE INDEX IF NOT EXISTS idx_user_identities_user ON user_identities (user_id);

-- Project share tokens: one row per active share link. Tokens are long
-- random strings, not guessable. No expiry on creation — the owner
-- revokes from the share dialog. Public read goes through the token.
CREATE TABLE IF NOT EXISTS user_project_shares (
    token      TEXT        PRIMARY KEY,
    project_id TEXT        NOT NULL REFERENCES user_projects(id) ON DELETE CASCADE,
    user_id    TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_project_shares_project
    ON user_project_shares (project_id, created_at DESC);

-- Project versions: a rolling history of saved snapshots, Pro/BYO only.
-- We keep up to 10 most-recent snapshots per project (trim on insert).
-- Snapshot data is the raw diagram JSON; we don't deduplicate or diff.
CREATE TABLE IF NOT EXISTS user_project_versions (
    id         TEXT        PRIMARY KEY,
    project_id TEXT        NOT NULL REFERENCES user_projects(id) ON DELETE CASCADE,
    user_id    TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      TEXT        NOT NULL,
    data       JSONB       NOT NULL,
    byte_size  INT         NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_project_versions_project
    ON user_project_versions (project_id, created_at DESC);

-- Activation events: a single (user_id, kind) row marks the first time
-- the user did the thing. Idempotent writes via ON CONFLICT DO NOTHING.
-- Kinds in use: 'signup', 'first_diagram', 'first_iac_export'.
CREATE TABLE IF NOT EXISTS activation_events (
    user_id     TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind        TEXT        NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, kind)
);

CREATE INDEX IF NOT EXISTS idx_activation_events_kind
    ON activation_events (kind, occurred_at);

-- Audit log. Append-only record of security-relevant events: plan changes,
-- role changes, share token mints/revokes, IaC exports, login + password
-- reset, webhook signature failures. The actor_user_id may be NULL for
-- system-originated events (e.g. webhook handlers run unauthenticated).
-- Metadata is free-form JSON for event-specific context (ip, provider, etc).
-- We never UPDATE or DELETE rows — DBA-level grants should reflect this.
CREATE TABLE IF NOT EXISTS audit_events (
    id            TEXT        PRIMARY KEY,
    occurred_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    event         TEXT        NOT NULL,
    actor_user_id TEXT        REFERENCES users(id) ON DELETE SET NULL,
    target_id     TEXT,
    ip            TEXT,
    user_agent    TEXT,
    metadata      JSONB       NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_audit_events_event_ts
    ON audit_events (event, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_events_actor_ts
    ON audit_events (actor_user_id, occurred_at DESC);

-- Support / feedback inbox. Reachable by anonymous users via /support, and
-- pre-filled for logged-in users. user_id is nullable so signed-out posters
-- can still reach us. Status is a tiny state machine: 'new' → 'open' →
-- 'closed'. We don't surface any of this on the user side — admin-only.
CREATE TABLE IF NOT EXISTS feedback (
    id          TEXT        PRIMARY KEY,
    user_id     TEXT        REFERENCES users(id) ON DELETE SET NULL,
    name        TEXT        NOT NULL,
    email       TEXT        NOT NULL,
    category    TEXT        NOT NULL DEFAULT 'general',
    message     TEXT        NOT NULL,
    status      TEXT        NOT NULL DEFAULT 'new',
    ip          TEXT,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feedback_status_created
    ON feedback (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_feedback_user
    ON feedback (user_id, created_at DESC);

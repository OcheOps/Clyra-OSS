-- Activation events: coarse funnel markers for measuring what actually
-- matters (signup → first-diagram → first-export). Each (user_id, kind)
-- is unique so recording is idempotent — the first time a user does the
-- thing is what we care about; subsequent calls ON CONFLICT DO NOTHING.

CREATE TABLE IF NOT EXISTS activation_events (
    user_id    TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind       TEXT        NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, kind)
);

CREATE INDEX IF NOT EXISTS idx_activation_events_kind
    ON activation_events (kind, occurred_at);

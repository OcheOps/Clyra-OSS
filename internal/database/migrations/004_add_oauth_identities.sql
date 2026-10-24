-- +migrate Up
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

CREATE TABLE IF NOT EXISTS user_identities (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider     TEXT NOT NULL,
    provider_sub TEXT NOT NULL,
    email        TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_sub)
);

CREATE INDEX IF NOT EXISTS idx_user_identities_user ON user_identities (user_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_user_identities_user;
DROP TABLE IF EXISTS user_identities;
-- Intentionally not re-adding NOT NULL on password_hash: existing OAuth-only
-- users would have NULL hashes and the migration would fail. Downgrade would
-- need a data-migration step to delete OAuth-only users first.

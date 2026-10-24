-- +migrate Up
CREATE TABLE IF NOT EXISTS invite_codes (
    id           TEXT PRIMARY KEY,
    code         TEXT NOT NULL UNIQUE,
    created_by   TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    used_by      TEXT REFERENCES users(id) ON DELETE SET NULL,
    used_at      TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invite_codes_code ON invite_codes (code);

-- +migrate Down
DROP TABLE IF EXISTS invite_codes;

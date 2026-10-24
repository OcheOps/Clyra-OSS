-- +migrate Up

-- Daily usage counters for quota enforcement
CREATE TABLE IF NOT EXISTS usage_records (
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    usage_date    DATE NOT NULL DEFAULT CURRENT_DATE,
    request_count INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, usage_date)
);

CREATE INDEX idx_usage_records_date ON usage_records (usage_date);

-- Granular usage log for observability (no sensitive data)
CREATE TABLE IF NOT EXISTS usage_log (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider   TEXT NOT NULL,
    success    BOOLEAN NOT NULL,
    latency_ms INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_usage_log_user_date ON usage_log (user_id, created_at);

-- +migrate Down
DROP TABLE IF EXISTS usage_log;
DROP TABLE IF EXISTS usage_records;

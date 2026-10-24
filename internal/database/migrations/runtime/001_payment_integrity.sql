-- Payment-provider identity and renewal idempotency.
ALTER TABLE payments ADD COLUMN IF NOT EXISTS checkout_id TEXT;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS recurring BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_checkout_id
    ON payments (checkout_id) WHERE checkout_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_order_id
    ON payments (order_id) WHERE order_id IS NOT NULL;

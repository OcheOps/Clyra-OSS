package payments

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// All column lists include `external_id` (provider-native cancel handle, e.g.
// Polar subscription UUID) and `order_id` (Polar order handle required by the
// refunds API). Paystack leaves both NULL — its ProviderRef does both jobs.
const paymentCols = `id, user_id, provider, provider_ref, checkout_id, recurring, external_id, order_id, email_token, amount_minor, currency, region, plan, period_days, status, created_at, verified_at`

func scanPayment(row pgx.Row, p *Payment) error {
	return row.Scan(
		&p.ID, &p.UserID, &p.Provider, &p.ProviderRef, &p.CheckoutID, &p.Recurring, &p.ExternalID, &p.OrderID, &p.EmailToken,
		&p.AmountMinor, &p.Currency, &p.Region, &p.Plan, &p.PeriodDays,
		&p.Status, &p.CreatedAt, &p.VerifiedAt,
	)
}

func (r *Repository) Create(ctx context.Context, p *Payment) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO payments
		  (id, user_id, provider, provider_ref, checkout_id, recurring, external_id, order_id, email_token, amount_minor, currency, region, plan, period_days, status, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		p.ID, p.UserID, p.Provider, p.ProviderRef, p.CheckoutID, p.Recurring, p.ExternalID, p.OrderID, p.EmailToken,
		p.AmountMinor, p.Currency, p.Region, p.Plan, p.PeriodDays, p.Status, p.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting payment: %w", err)
	}
	return nil
}

// RecordRenewalAndExtend records a renewal and extends the user's entitlement
// in one transaction. Unique provider identifiers make concurrent replays a
// no-op instead of granting the same paid period twice.
func (r *Repository) RecordRenewalAndExtend(ctx context.Context, p *Payment) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("beginning renewal transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`INSERT INTO payments
		  (id, user_id, provider, provider_ref, checkout_id, recurring, external_id, order_id, email_token, amount_minor, currency, region, plan, period_days, status, created_at, verified_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,'paid',$15,NOW())
		 ON CONFLICT DO NOTHING`,
		p.ID, p.UserID, p.Provider, p.ProviderRef, p.CheckoutID, p.Recurring,
		p.ExternalID, p.OrderID, p.EmailToken, p.AmountMinor, p.Currency,
		p.Region, p.Plan, p.PeriodDays, p.CreatedAt,
	)
	if err != nil {
		return false, fmt.Errorf("inserting renewal: %w", err)
	}
	if tag.RowsAffected() == 0 {
		if err := tx.Commit(ctx); err != nil {
			return false, fmt.Errorf("committing duplicate renewal: %w", err)
		}
		return false, nil
	}

	tag, err = tx.Exec(ctx,
		`UPDATE users
		 SET plan = $1,
		     plan_expiry = GREATEST(COALESCE(plan_expiry, NOW()), NOW()) + ($2 * INTERVAL '1 day'),
		     updated_at = NOW()
		 WHERE id = $3`,
		p.Plan, p.PeriodDays, p.UserID,
	)
	if err != nil {
		return false, fmt.Errorf("extending renewal entitlement: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return false, fmt.Errorf("extending renewal entitlement: user not found")
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("committing renewal: %w", err)
	}
	return true, nil
}

func (r *Repository) SetProviderMetadata(ctx context.Context, ref, externalID, orderID, emailToken string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET
		 external_id = COALESCE(NULLIF(external_id, ''), NULLIF($1, '')),
		 order_id = COALESCE(NULLIF(order_id, ''), NULLIF($2, '')),
		 email_token = COALESCE(NULLIF(email_token, ''), NULLIF($3, ''))
		 WHERE provider_ref = $4`,
		externalID, orderID, emailToken, ref,
	)
	if err != nil {
		return fmt.Errorf("setting provider metadata: %w", err)
	}
	return nil
}

// CompleteInitial applies the initial payment state and entitlement together.
// It also repairs provider metadata on an already-paid row without granting a
// second entitlement when browser verification races the webhook.
func (r *Repository) CompleteInitial(ctx context.Context, p *Payment, status Status, externalID, orderID, emailToken string) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("beginning payment completion: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var current Status
	if err := tx.QueryRow(ctx,
		`SELECT status FROM payments WHERE provider_ref = $1 FOR UPDATE`, p.ProviderRef,
	).Scan(&current); err != nil {
		return false, fmt.Errorf("locking payment: %w", err)
	}
	if current == StatusFailed && status == StatusPaid {
		return false, fmt.Errorf("payment is already failed")
	}

	if current != StatusPaid && status == StatusPaid {
		expiry := time.Now().UTC().AddDate(0, 0, p.PeriodDays)
		tag, err := tx.Exec(ctx,
			`UPDATE users SET plan = $1, plan_expiry = $2, updated_at = NOW() WHERE id = $3`,
			p.Plan, expiry, p.UserID,
		)
		if err != nil {
			return false, fmt.Errorf("granting initial entitlement: %w", err)
		}
		if tag.RowsAffected() != 1 {
			return false, fmt.Errorf("granting initial entitlement: user not found")
		}
	}

	_, err = tx.Exec(ctx,
		`UPDATE payments SET
		 status = CASE WHEN status = 'paid' THEN status ELSE $1 END,
		 verified_at = CASE WHEN $1 = 'paid' THEN COALESCE(verified_at, NOW()) ELSE verified_at END,
		 external_id = COALESCE(NULLIF(external_id, ''), NULLIF($2, '')),
		 order_id = COALESCE(NULLIF(order_id, ''), NULLIF($3, '')),
		 email_token = COALESCE(NULLIF(email_token, ''), NULLIF($4, ''))
		 WHERE provider_ref = $5`,
		status, externalID, orderID, emailToken, p.ProviderRef,
	)
	if err != nil {
		return false, fmt.Errorf("completing payment: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("committing payment completion: %w", err)
	}
	return current != StatusPaid && status == StatusPaid, nil
}

// SetEmailToken persists the Paystack subscription email_token captured on
// the initial charge.success or subscription.create webhook. Cancel needs
// it alongside subscription_code. No-op when token is empty.
func (r *Repository) SetEmailToken(ctx context.Context, ref, token string) error {
	if token == "" {
		return nil
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET email_token = $1 WHERE provider_ref = $2 AND (email_token IS NULL OR email_token = '')`,
		token, ref,
	)
	if err != nil {
		return fmt.Errorf("setting email token: %w", err)
	}
	return nil
}

func (r *Repository) GetByRef(ctx context.Context, ref string) (*Payment, error) {
	var p Payment
	err := scanPayment(
		r.pool.QueryRow(ctx, `SELECT `+paymentCols+` FROM payments WHERE provider_ref = $1`, ref),
		&p,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting payment: %w", err)
	}
	return &p, nil
}

func (r *Repository) MarkStatus(ctx context.Context, ref string, status Status) error {
	var verifiedAt *time.Time
	if status == StatusPaid {
		t := time.Now().UTC()
		verifiedAt = &t
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET status = $1, verified_at = COALESCE(verified_at, $2) WHERE provider_ref = $3`,
		status, verifiedAt, ref,
	)
	if err != nil {
		return fmt.Errorf("marking payment status: %w", err)
	}
	return nil
}

// SetExternalID persists the provider-native handle (e.g. Polar subscription
// UUID) we need to cancel / refund later. No-op when externalID is empty.
func (r *Repository) SetExternalID(ctx context.Context, ref, externalID string) error {
	if externalID == "" {
		return nil
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET external_id = $1 WHERE provider_ref = $2 AND (external_id IS NULL OR external_id = '')`,
		externalID, ref,
	)
	if err != nil {
		return fmt.Errorf("setting external id: %w", err)
	}
	return nil
}

// SetOrderID persists the Polar order_id captured on verify. Needed because
// /v1/refunds/ keys on order_id specifically. No-op when orderID is empty
// (Paystack, or a verify that came back before Polar created the order row).
func (r *Repository) SetOrderID(ctx context.Context, ref, orderID string) error {
	if orderID == "" {
		return nil
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET order_id = $1 WHERE provider_ref = $2 AND (order_id IS NULL OR order_id = '')`,
		orderID, ref,
	)
	if err != nil {
		return fmt.Errorf("setting order id: %w", err)
	}
	return nil
}

// CountPendingToday returns how many pending payment rows a user has created
// in the last 24 hours — used to rate-limit checkout spam.
func (r *Repository) CountPendingToday(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM payments
		 WHERE user_id = $1 AND status = 'pending' AND created_at > NOW() - INTERVAL '24 hours'`,
		userID,
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("counting pending payments: %w", err)
	}
	return n, nil
}

// GetLatestPaidByUser returns the user's most recent paid (non-refunded) payment,
// or nil if they have none. Used by the cancel flow to decide whether a refund
// is still eligible.
func (r *Repository) GetLatestPaidByUser(ctx context.Context, userID string) (*Payment, error) {
	var p Payment
	err := scanPayment(
		r.pool.QueryRow(ctx,
			`SELECT `+paymentCols+` FROM payments
			 WHERE user_id = $1 AND status = 'paid'
			 ORDER BY verified_at DESC NULLS LAST, created_at DESC
			 LIMIT 1`,
			userID,
		), &p,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting latest paid payment: %w", err)
	}
	return &p, nil
}

// GetPaidBySubscription finds the most recent paid payment for a given
// provider subscription (Polar subscription UUID stored in external_id).
// Used by the renewal webhook to identify the user + plan + period a
// recurring order.paid belongs to. Returns nil if we've never seen this
// subscription (e.g. webhook for a checkout we didn't create).
func (r *Repository) GetPaidBySubscription(ctx context.Context, subscriptionID string) (*Payment, error) {
	if subscriptionID == "" {
		return nil, nil
	}
	var p Payment
	err := scanPayment(
		r.pool.QueryRow(ctx,
			`SELECT `+paymentCols+` FROM payments
			 WHERE external_id = $1 AND status = 'paid'
			 ORDER BY verified_at DESC NULLS LAST, created_at DESC
			 LIMIT 1`,
			subscriptionID,
		), &p,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting paid payment by subscription: %w", err)
	}
	return &p, nil
}

// GetByOrderID finds a payment by Polar order_id. Used to dedupe renewal
// webhooks — if we've already recorded this order (whether from a prior
// delivery or a race), the caller can short-circuit without extending the
// user's plan a second time.
func (r *Repository) GetByOrderID(ctx context.Context, orderID string) (*Payment, error) {
	if orderID == "" {
		return nil, nil
	}
	var p Payment
	err := scanPayment(
		r.pool.QueryRow(ctx,
			`SELECT `+paymentCols+` FROM payments WHERE order_id = $1 LIMIT 1`,
			orderID,
		), &p,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting payment by order id: %w", err)
	}
	return &p, nil
}

// ExpireStalePending flips any payment still in 'pending' older than the given
// cutoff to 'failed'. Orphaned rows (user bailed out of checkout, Polar
// metadata-verify never ran, etc.) otherwise block fresh checkouts because
// the service refuses to initialize a new payment while one is pending for
// the same plan. Returns the number of rows updated.
func (r *Repository) ExpireStalePending(ctx context.Context, olderThan time.Duration) (int64, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE payments SET status = 'failed'
		 WHERE status = 'pending' AND created_at < NOW() - $1::interval`,
		olderThan.String(),
	)
	if err != nil {
		return 0, fmt.Errorf("expiring stale pending payments: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]Payment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+paymentCols+` FROM payments WHERE user_id = $1 ORDER BY created_at DESC LIMIT 100`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing payments: %w", err)
	}
	defer rows.Close()
	var out []Payment
	for rows.Next() {
		var p Payment
		if err := scanPayment(rows, &p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

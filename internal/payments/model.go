// Package payments wires one-time plan upgrades through an abstracted payment
// provider (Paystack today; Stripe / AWS Marketplace later).
//
// Design goals:
//   - No UI in the current product surface — endpoints exist but are not
//     exposed in the landing page. This lets billing be flipped on later
//     by wiring buttons, nothing else.
//   - Mock mode when no secret key is configured: initialize + verify return
//     deterministic success so devs can exercise the full upgrade flow.
//   - Provider-agnostic DB records so Stripe can slot in as a second provider.
package payments

import "time"

// Status tracks the lifecycle of a single payment record.
type Status string

const (
	StatusPending  Status = "pending"
	StatusPaid     Status = "paid"
	StatusFailed   Status = "failed"
	StatusRefunded Status = "refunded"
)

// Payment is the canonical record persisted in the payments table.
type Payment struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	Provider    string  `json:"provider"`
	ProviderRef string  `json:"provider_ref"`
	CheckoutID  *string `json:"checkout_id,omitempty"`
	Recurring   bool    `json:"recurring"`
	// ExternalID is the provider-native subscription/recurring handle:
	// Polar's subscription UUID, or Paystack's subscription_code once we
	// initialize checkouts via /transaction/initialize?plan=... . NULL
	// for Paystack one-time charges made before subscriptions existed.
	ExternalID *string `json:"external_id,omitempty"`
	// OrderID is Polar's order record (the only handle Polar's /v1/refunds/
	// API accepts). NULL for Paystack — its ProviderRef already refunds.
	OrderID *string `json:"order_id,omitempty"`
	// EmailToken is Paystack's per-subscription token required by
	// /subscription/disable (paired with subscription_code in ExternalID).
	// NULL for Polar and pre-subscription Paystack charges.
	EmailToken  *string    `json:"email_token,omitempty"`
	AmountMinor int        `json:"amount_minor"`
	Currency    string     `json:"currency"`
	Region      string     `json:"region"`
	Plan        string     `json:"plan"`
	PeriodDays  int        `json:"period_days"`
	Status      Status     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	VerifiedAt  *time.Time `json:"verified_at,omitempty"`
}

// InitRequest is what the frontend sends to start a checkout.
type InitRequest struct {
	Plan   string `json:"plan"`   // "pro" | "byo"
	Region string `json:"region"` // "NG" | "" (global)
}

// InitResponse is what the frontend receives to redirect the user.
type InitResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	Reference        string `json:"reference"`
	// ExternalID is the provider's native checkout ID when distinct from
	// Reference (e.g. Polar's UUID). The frontend stashes this so it can
	// call verify with the provider's ID even when the provider doesn't
	// append it to success_url on redirect.
	ExternalID string `json:"external_id,omitempty"`
	Provider   string `json:"provider,omitempty"`
	Mock       bool   `json:"mock,omitempty"`
}

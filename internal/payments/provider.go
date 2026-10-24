package payments

import (
	"context"
	"errors"
)

// ProviderResult represents the outcome of a provider's initialize/verify call.
type ProviderResult struct {
	AuthorizationURL string // for initialize
	Reference        string // provider-assigned reference (usually our own, echoed back)
	// ExternalID is the provider's native ID when distinct from Reference —
	// Polar's checkout UUID is the canonical example (Polar doesn't append
	// it to custom success_url, so the frontend needs it to active-verify).
	// Empty for Paystack — its reference already is its ID.
	ExternalID string
	// OrderID is the provider's order-record handle, distinct from both
	// Reference and ExternalID. Polar's refunds API only accepts order_id
	// (not subscription_id, not checkout_id), so we capture it on verify
	// and persist it for later use. Empty for Paystack.
	OrderID     string
	AmountMinor int
	Currency    string
	Plan        string
	Recurring   bool
	Status      Status // for verify
	Mock        bool
}

// ProviderInitInput carries what any provider needs to start a checkout.
type ProviderInitInput struct {
	Email       string
	AmountMinor int
	Currency    string
	Plan        string
	Region      string
	Reference   string // locally-minted, passed to provider when supported
}

// Provider is implemented by each payment backend.
type Provider interface {
	Name() string
	Initialize(ctx context.Context, in ProviderInitInput) (ProviderResult, error)
	Verify(ctx context.Context, reference string) (ProviderResult, error)
	// Refund issues a full refund against a completed transaction. Returns
	// ErrRefundUnavailable if the provider doesn't support programmatic refunds.
	Refund(ctx context.Context, reference string) error
}

// ErrRefundUnavailable signals a provider that can't issue refunds via API.
var ErrRefundUnavailable = errors.New("provider does not support programmatic refunds")

// ErrUnsupported is returned by providers that can't handle a given currency/region.
var ErrUnsupported = errors.New("payment currency not supported by provider")

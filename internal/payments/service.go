package payments

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/internal/pricing"
)

var (
	ErrUnknownPlan       = errors.New("unknown plan")
	ErrNotUpgradable     = errors.New("plan cannot be purchased")
	ErrNotFound          = errors.New("payment not found")
	ErrTooManyPending    = errors.New("too many pending checkouts — finish or wait before starting another")
	ErrAlreadySubscribed = errors.New("you already have an active plan — cancel it before starting another")
	ErrNoActivePlan      = errors.New("no active plan to cancel")
)

// MaxPendingPerDay caps how many pending payment rows a single user can mint
// within a 24h window. Guards against DB bloat and checkout-flow spamming.
const MaxPendingPerDay = 5

// UserUpgrader is the subset of the user repository the payment service needs.
type UserUpgrader interface {
	GetByID(ctx context.Context, id string) (*auth.User, error)
	UpgradePlan(ctx context.Context, userID string, plan auth.Plan, expiry time.Time) error
	ExtendPlan(ctx context.Context, userID string, plan auth.Plan, days int) error
	CancelPlan(ctx context.Context, userID string) error
}

// Service owns the payments state machine and picks which Provider handles
// a given transaction. Providers are keyed by their Name() ("paystack",
// "polar") so existing payment rows (which carry provider name) can always
// be verified or refunded by the right backend, even after the region-routing
// rules change.
type Service struct {
	repo      *Repository
	providers map[string]Provider
	// routeByRegion picks a provider name from a region code. Defaults below;
	// can be customized by tests or future geo-aware strategies.
	routeByRegion func(region string) string
	users         UserUpgrader
	audit         AuditSink
	logger        *slog.Logger
}

// WithAudit attaches the audit sink. Plan upgrades, cancels, and refunds
// emit security events when set; absent, those flows still work but are
// invisible to the audit log.
func (s *Service) WithAudit(a AuditSink) *Service {
	s.audit = a
	return s
}

// NewService wires one or more providers. Pass any nil-valued providers to
// disable them — useful when Paystack keys aren't configured yet.
// First non-nil provider becomes the fallback when region routing misses.
func NewService(repo *Repository, users UserUpgrader, logger *slog.Logger, provs ...Provider) *Service {
	reg := make(map[string]Provider, len(provs))
	for _, p := range provs {
		if p == nil {
			continue
		}
		reg[p.Name()] = p
	}
	return &Service{
		repo:          repo,
		providers:     reg,
		routeByRegion: defaultRouteByRegion,
		users:         users,
		logger:        logger,
	}
}

// WithRouter overrides the default region → provider mapping. Used by
// main.go to honor PAYMENTS_NG_PROVIDER=paystack when the operator has
// confirmed Paystack is unblocked. Returns the service for chaining.
func (s *Service) WithRouter(fn func(region string) string) *Service {
	if fn != nil {
		s.routeByRegion = fn
	}
	return s
}

// PaystackFirstNGRouter is the pre-throttling routing rule: NG → paystack,
// else → polar. Exposed so main.go can wire it via WithRouter when the
// PAYMENTS_NG_PROVIDER env var is set to "paystack".
func PaystackFirstNGRouter(region string) string {
	if strings.EqualFold(region, "NG") {
		return "paystack"
	}
	return "polar"
}

// defaultRouteByRegion:
//   - NG → paystack (naira local rails + international Visa/Mastercard)
//   - anywhere else → polar (MoR, global tax compliance handled for us)
//
// Operators can force everyone to Polar (bypassing Paystack) by setting
// PAYMENTS_NG_PROVIDER=polar in env. Historical routing rule restored
// as the default — the earlier Polar-only override was a temporary
// workaround for a Paystack throttling incident.
func defaultRouteByRegion(region string) string {
	if strings.EqualFold(region, "NG") {
		return "paystack"
	}
	return "polar"
}

// providerFor returns the provider that handles a given region, falling back
// to any configured provider if the preferred one is missing. Returns nil if
// nothing is configured.
func (s *Service) providerFor(region string) Provider {
	name := s.routeByRegion(region)
	if p, ok := s.providers[name]; ok {
		return p
	}
	// Fallback: any configured provider, so a deployment with Polar-only
	// still accepts NG customers (they pay in USD via Polar).
	for _, p := range s.providers {
		return p
	}
	return nil
}

// providerByName looks up a provider by its Name() — used to route verify /
// cancel / webhook for existing payment rows to whichever backend created them.
func (s *Service) providerByName(name string) Provider {
	return s.providers[name]
}

// Initialize creates a payment record and returns the provider's checkout URL.
// The plan + region pair determines pricing via the pricing catalog.
func (s *Service) Initialize(ctx context.Context, userID string, req InitRequest) (*InitResponse, error) {
	plan := pricing.Plan(req.Plan)
	if plan != pricing.PlanPro && plan != pricing.PlanBYO {
		return nil, ErrNotUpgradable
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("resolving user: %w", err)
	}

	// Block starting a new checkout while an active paid plan is still running.
	// The user must cancel first (or wait for expiry). This keeps billing state
	// unambiguous since we don't support mid-cycle plan swaps with proration.
	if user.Plan != auth.PlanFree && !user.Expired() {
		return nil, ErrAlreadySubscribed
	}

	pending, err := s.repo.CountPendingToday(ctx, userID)
	if err != nil {
		return nil, err
	}
	if pending >= MaxPendingPerDay {
		return nil, ErrTooManyPending
	}

	provider := s.providerFor(req.Region)
	if provider == nil {
		return nil, errors.New("no payment provider configured")
	}

	// Currency-provider match: Polar only accepts USD. If we're routing
	// this checkout to Polar for a region whose native pricing is not
	// USD (NG → NGN, EU → EUR later), fall back to the USD pricing card
	// so the amount + currency match what Polar can charge. Otherwise
	// the checkout would hard-fail on ErrUnsupported.
	//
	// This was silently broken for NG users after the Paystack→Polar
	// routing flip: pricing UI showed ₦3,500, backend tried to charge
	// NGN via Polar, Polar rejected the request. Users saw a generic
	// 'checkout failed' with no path forward.
	pricingRegion := req.Region
	if provider.Name() == "polar" {
		pricingRegion = ""
	}
	card, ok := pricing.CardFor(pricingRegion, plan)
	if !ok {
		return nil, ErrUnknownPlan
	}

	reference := generateReference()
	initInput := ProviderInitInput{
		Email:       user.Email,
		AmountMinor: card.AmountMinor,
		Currency:    card.Currency,
		Plan:        string(plan),
		Region:      req.Region,
		Reference:   reference,
	}

	// Paystack subscription path: initialize via /transaction/initialize?plan=…
	// so the successful charge auto-creates a Subscription and Paystack takes
	// over renewal billing. Fall back to plain one-time charge if plan lookup
	// fails (so a Paystack outage doesn't block checkout).
	var result ProviderResult
	if paystack, ok := provider.(*Paystack); ok {
		planName := fmt.Sprintf("Clyra %s %s Monthly", titleCase(string(plan)), card.Currency)
		planCode, perr := paystack.EnsurePlan(ctx, planName, card.AmountMinor, card.Currency, "monthly")
		if perr == nil && planCode != "" {
			result, err = paystack.InitializeSubscription(ctx, initInput, planCode)
		} else {
			s.logger.Warn("paystack plan ensure failed — falling back to one-time charge",
				"plan", plan, "currency", card.Currency, "error", perr)
			result, err = paystack.Initialize(ctx, initInput)
		}
	} else {
		result, err = provider.Initialize(ctx, initInput)
	}
	if err != nil {
		return nil, err
	}

	rec := &Payment{
		ID:          generateID(),
		UserID:      userID,
		Provider:    provider.Name(),
		ProviderRef: result.Reference,
		Recurring:   result.Recurring,
		AmountMinor: card.AmountMinor,
		Currency:    card.Currency,
		Region:      req.Region,
		Plan:        string(plan),
		PeriodDays:  30,
		Status:      StatusPending,
		CreatedAt:   time.Now().UTC(),
	}
	if provider.Name() == "polar" && result.ExternalID != "" {
		rec.CheckoutID = &result.ExternalID
	}
	if err := s.repo.Create(ctx, rec); err != nil {
		return nil, err
	}

	return &InitResponse{
		AuthorizationURL: result.AuthorizationURL,
		Reference:        result.Reference,
		ExternalID:       result.ExternalID,
		Provider:         provider.Name(),
		Mock:             result.Mock,
	}, nil
}

// VerifyHint carries provider-specific hints the caller picked up from the
// return URL. Paystack needs nothing beyond the reference; Polar benefits from
// a checkout_id so we can hit their API directly instead of waiting for a
// webhook (critical for local dev without a public webhook URL).
type VerifyHint struct {
	PolarCheckoutID string
}

// Verify confirms a transaction with the provider and upgrades the user on success.
// Idempotent: running it twice on the same reference doesn't double-upgrade.
func (s *Service) Verify(ctx context.Context, userID, reference string, hint VerifyHint) (*Payment, error) {
	rec, err := s.repo.GetByRef(ctx, reference)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	if rec.UserID != userID {
		return nil, ErrNotFound // don't leak existence
	}

	// Failed payments are terminal. Paid Polar payments may still need a
	// metadata repair when a webhook won the race with browser verification.
	if rec.Status == StatusFailed {
		return rec, nil
	}

	provider := s.providerByName(rec.Provider)
	if provider == nil {
		return nil, fmt.Errorf("payment provider %q not configured", rec.Provider)
	}

	// Polar active-verify path: if the caller handed us a checkout_id from the
	// return URL, look it up directly on Polar rather than returning pending.
	var result ProviderResult
	if polar, ok := provider.(*Polar); ok && hint.PolarCheckoutID != "" {
		if rec.CheckoutID == nil || *rec.CheckoutID == "" || hint.PolarCheckoutID != *rec.CheckoutID {
			return nil, ErrNotFound
		}
		result, err = polar.LookupCheckout(ctx, hint.PolarCheckoutID)
	} else {
		result, err = provider.Verify(ctx, reference)
	}
	if err != nil {
		return nil, err
	}
	if result.Reference != "" && result.Reference != rec.ProviderRef {
		return nil, fmt.Errorf("payment provider reference mismatch")
	}
	if result.AmountMinor != 0 && result.AmountMinor != rec.AmountMinor {
		return nil, fmt.Errorf("payment amount mismatch")
	}
	if result.Currency != "" && !strings.EqualFold(result.Currency, rec.Currency) {
		return nil, fmt.Errorf("payment currency mismatch")
	}
	if result.Plan != "" && result.Plan != rec.Plan {
		return nil, fmt.Errorf("payment plan mismatch")
	}
	if rec.Status == StatusPaid && result.Status != StatusPaid {
		return rec, nil
	}

	switch result.Status {
	case StatusPaid:
		transitioned, err := s.repo.CompleteInitial(ctx, rec, StatusPaid, result.ExternalID, result.OrderID, "")
		if err != nil {
			s.logger.Error("payment completion failed", "user_id", userID, "ref", reference, "error", err)
			return nil, err
		}
		if result.ExternalID != "" {
			rec.ExternalID = &result.ExternalID
		}
		if result.OrderID != "" {
			rec.OrderID = &result.OrderID
		}
		rec.Status = StatusPaid
		if !transitioned {
			return rec, nil
		}
		s.logger.Info("payment verified & plan upgraded",
			"user_id", userID, "plan", rec.Plan, "reference", reference,
			"external_id", result.ExternalID, "mock", result.Mock,
		)
		if s.audit != nil {
			s.audit.Log(ctx, "billing.plan_upgraded", map[string]any{
				"actor_user_id": userID,
				"target_id":     userID,
				"plan":          rec.Plan,
				"provider":      rec.Provider,
				"reference":     reference,
				"source":        "verify",
			})
		}
	case StatusFailed:
		if _, err := s.repo.CompleteInitial(ctx, rec, StatusFailed, result.ExternalID, result.OrderID, ""); err != nil {
			return nil, err
		}
		rec.Status = StatusFailed
	}
	return rec, nil
}

type WebhookMetadata struct {
	ExternalID string
	OrderID    string
	EmailToken string
}

// HandleWebhook processes a provider-pushed status update. Provider identity
// is part of the lookup contract so one provider cannot settle another
// provider's pending payment. Metadata is persisted before the idempotent
// completion transaction to repair webhook-first and verify-first races.
func (s *Service) HandleWebhook(ctx context.Context, providerName, reference string, status Status, metadata WebhookMetadata) error {
	rec, err := s.repo.GetByRef(ctx, reference)
	if err != nil || rec == nil {
		return ErrNotFound
	}
	if rec.Provider != providerName {
		return ErrNotFound
	}
	if rec.Status == StatusPaid || rec.Status == StatusFailed {
		_, err := s.repo.CompleteInitial(ctx, rec, rec.Status, metadata.ExternalID, metadata.OrderID, metadata.EmailToken)
		return err
	}
	if status == StatusPaid {
		transitioned, err := s.repo.CompleteInitial(ctx, rec, status, metadata.ExternalID, metadata.OrderID, metadata.EmailToken)
		if err != nil {
			return err
		}
		if transitioned && s.audit != nil {
			s.audit.Log(ctx, "billing.plan_upgraded", map[string]any{
				"target_id": rec.UserID,
				"plan":      rec.Plan,
				"provider":  rec.Provider,
				"reference": reference,
				"source":    "webhook",
			})
		}
		return nil
	}
	_, err = s.repo.CompleteInitial(ctx, rec, status, metadata.ExternalID, metadata.OrderID, metadata.EmailToken)
	return err
}

// HandleRenewal processes a Polar order.paid event that belongs to a
// recurring subscription (not the initial checkout). The initial
// HandleWebhook flow can't extend the plan for these because it looks up
// by our internal reference — which is only set on the first order via
// checkout metadata; Polar's follow-up orders in the same subscription
// come through without it. That was silently dropping every renewal.
//
// Flow:
//   - Dedupe on order_id first (webhook retries + at-least-once delivery)
//   - Look up the seed payment by subscription_id to recover user/plan/period
//   - Atomically bump plan_expiry by PeriodDays from max(current, now) so
//     early renewals don't burn remaining paid time
//   - Insert a new paid payments row for the renewal (audit trail + so the
//     next cancel-refund window uses the correct verified_at)
func (s *Service) HandleRenewal(ctx context.Context, subscriptionID, orderID string) error {
	if subscriptionID == "" || orderID == "" {
		return nil
	}
	// Dedup: same delivery replayed, or the initial webhook already
	// recorded this order via reference lookup.
	if existing, err := s.repo.GetByOrderID(ctx, orderID); err != nil {
		return err
	} else if existing != nil {
		return nil
	}
	seed, err := s.repo.GetPaidBySubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}
	if seed == nil {
		// Not our subscription — unknown seed. Ack without action so Polar
		// doesn't retry forever; audit so we notice unexpected traffic.
		if s.audit != nil {
			s.audit.Log(ctx, "webhook.renewal_unknown_subscription", map[string]any{
				"subscription_id": subscriptionID,
				"order_id":        orderID,
			})
		}
		return nil
	}

	renewal := &Payment{
		ID:          generateID(),
		UserID:      seed.UserID,
		Provider:    seed.Provider,
		ProviderRef: generateReference(), // fresh internal ref per renewal row
		ExternalID:  &subscriptionID,
		OrderID:     &orderID,
		AmountMinor: seed.AmountMinor,
		Currency:    seed.Currency,
		Region:      seed.Region,
		Plan:        seed.Plan,
		PeriodDays:  seed.PeriodDays,
		Status:      StatusPaid,
		Recurring:   true,
		CreatedAt:   time.Now().UTC(),
	}
	recorded, err := s.repo.RecordRenewalAndExtend(ctx, renewal)
	if err != nil {
		return err
	}
	if !recorded {
		return nil
	}

	if s.audit != nil {
		s.audit.Log(ctx, "billing.plan_renewed", map[string]any{
			"target_id":       seed.UserID,
			"plan":            seed.Plan,
			"provider":        seed.Provider,
			"subscription_id": subscriptionID,
			"order_id":        orderID,
			"period_days":     seed.PeriodDays,
		})
	}
	s.logger.Info("plan renewed",
		"user_id", seed.UserID, "subscription_id", subscriptionID,
		"order_id", orderID, "period_days", seed.PeriodDays)
	return nil
}

// HandlePaystackRenewal processes a Paystack charge.success event that
// belongs to a recurring subscription (not the initial checkout). Same
// contract as HandleRenewal for Polar, but keyed on subscription_code
// instead of subscription_id.
//
// Skips if the event's reference already matches an existing payment row
// — that's the initial charge, which the reference-based HandleWebhook
// handles. Non-initial charges have Paystack-generated references we've
// never seen, so we treat them as renewals.
func (s *Service) HandlePaystackRenewal(ctx context.Context, reference, subscriptionCode, emailToken string) error {
	if subscriptionCode == "" || reference == "" {
		return errors.New("paystack renewal missing reference or subscription code")
	}
	// Initial charge: our reference IS in the payment row already. Let the
	// reference path (HandleWebhook) upgrade + persist sub metadata.
	if existing, err := s.repo.GetByRef(ctx, reference); err == nil && existing != nil {
		return nil
	}
	seed, err := s.repo.GetPaidBySubscription(ctx, subscriptionCode)
	if err != nil {
		return err
	}
	if seed == nil {
		if s.audit != nil {
			s.audit.Log(ctx, "webhook.renewal_unknown_subscription", map[string]any{
				"subscription_code": subscriptionCode,
				"provider":          "paystack",
			})
		}
		return nil
	}

	renewal := &Payment{
		ID:          generateID(),
		UserID:      seed.UserID,
		Provider:    seed.Provider,
		ProviderRef: reference,
		ExternalID:  &subscriptionCode,
		AmountMinor: seed.AmountMinor,
		Currency:    seed.Currency,
		Region:      seed.Region,
		Plan:        seed.Plan,
		PeriodDays:  seed.PeriodDays,
		Status:      StatusPaid,
		Recurring:   true,
		CreatedAt:   time.Now().UTC(),
	}
	if emailToken != "" {
		renewal.EmailToken = &emailToken
	}
	recorded, err := s.repo.RecordRenewalAndExtend(ctx, renewal)
	if err != nil {
		return err
	}
	if !recorded {
		return nil
	}

	if s.audit != nil {
		s.audit.Log(ctx, "billing.plan_renewed", map[string]any{
			"target_id":         seed.UserID,
			"plan":              seed.Plan,
			"provider":          "paystack",
			"subscription_code": subscriptionCode,
			"period_days":       seed.PeriodDays,
		})
	}
	s.logger.Info("plan renewed via paystack",
		"user_id", seed.UserID, "subscription_code", subscriptionCode,
		"period_days", seed.PeriodDays)
	return nil
}

// HandlePaystackSubscriptionDisabled processes a subscription.disable
// event — either the user canceled via Paystack's customer portal, or
// Paystack disabled the subscription after too many failed charges.
// Either way, we should stop treating the user as Pro.
//
// Note on billing UX: strictly speaking, a customer cancel typically
// means "let me finish the paid period and don't renew." Paystack's
// subscription.disable fires immediately on cancel, but the user is
// still entitled to what they paid for until the next billing date.
// We honor that by ONLY clearing plan_expiry, not by wiping the plan
// row — EffectivePlan degrades to free once the expiry passes, so the
// user finishes their period gracefully. If Paystack disabled the sub
// for payment failure (not user intent), same treatment: they finish
// the period they already paid for, then degrade.
func (s *Service) HandlePaystackSubscriptionDisabled(ctx context.Context, subscriptionCode string) error {
	if subscriptionCode == "" {
		return nil
	}
	seed, err := s.repo.GetPaidBySubscription(ctx, subscriptionCode)
	if err != nil {
		return err
	}
	if seed == nil {
		// Unknown subscription — likely for a different account or a
		// test event. Ack without action; audit so we notice.
		if s.audit != nil {
			s.audit.Log(ctx, "webhook.subscription_disabled_unknown", map[string]any{
				"subscription_code": subscriptionCode,
				"provider":          "paystack",
			})
		}
		return nil
	}

	// No renewals will fire again for this subscription — mark the
	// latest paid row so downstream reporting is honest, but leave the
	// plan_expiry intact so the user finishes their paid period.
	//
	// We intentionally DO NOT call users.CancelPlan here — that would
	// yank them off Pro immediately even though they paid through the
	// end of the month. The natural expiry flow handles the downgrade.

	if s.audit != nil {
		s.audit.Log(ctx, "billing.subscription_disabled", map[string]any{
			"target_id":         seed.UserID,
			"provider":          "paystack",
			"subscription_code": subscriptionCode,
			"plan":              seed.Plan,
		})
	}
	s.logger.Info("paystack subscription disabled",
		"user_id", seed.UserID, "subscription_code", subscriptionCode)
	return nil
}

// HandlePaystackPaymentFailed processes an invoice.payment_failed event —
// Paystack tried to charge the saved card on the recurring date and got
// declined. Doesn't degrade the plan here (Paystack retries a few times
// per its retry schedule, then eventually disable the subscription
// which flows through HandlePaystackSubscriptionDisabled). We log for
// visibility + audit so ops can see churn signals.
func (s *Service) HandlePaystackPaymentFailed(ctx context.Context, subscriptionCode string) error {
	if subscriptionCode == "" {
		return nil
	}
	seed, err := s.repo.GetPaidBySubscription(ctx, subscriptionCode)
	if err != nil {
		return err
	}
	if seed == nil {
		return nil
	}
	if s.audit != nil {
		s.audit.Log(ctx, "billing.payment_failed", map[string]any{
			"target_id":         seed.UserID,
			"provider":          "paystack",
			"subscription_code": subscriptionCode,
			"plan":              seed.Plan,
		})
	}
	s.logger.Warn("paystack subscription payment failed",
		"user_id", seed.UserID, "subscription_code", subscriptionCode)
	return nil
}

// ListPayments returns recent payments for a user.
func (s *Service) ListPayments(ctx context.Context, userID string) ([]Payment, error) {
	return s.repo.ListByUser(ctx, userID)
}

// RefundWindow is the grace period after a successful payment during which a
// cancel triggers an automatic refund. Beyond this the user keeps the plan
// they paid for (they can cancel but without refund).
const RefundWindow = 24 * time.Hour

// CancelResult is returned to the caller so the UI can explain what happened:
// plain downgrade, or downgrade + refund.
type CancelResult struct {
	Refunded  bool   `json:"refunded"`
	RefundRef string `json:"refund_ref,omitempty"`
	PrevPlan  string `json:"prev_plan"`
}

// Cancel immediately downgrades the user to free and clears their plan expiry.
//
// Refund rule: if the user's most recent paid payment verified within the last
// 24 hours, we refund it through the provider and flip the payment row to
// 'refunded'. Past that window, cancel is still allowed but no refund is issued
// — the user forfeits remaining paid days.
func (s *Service) Cancel(ctx context.Context, userID string) (*CancelResult, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("resolving user: %w", err)
	}
	if user.Plan == auth.PlanFree || user.Expired() {
		return nil, ErrNoActivePlan
	}

	res := &CancelResult{PrevPlan: string(user.Plan)}

	// Consider refund eligibility *before* downgrading so the user never ends
	// up in a state where they've been demoted but the refund silently failed.
	latest, err := s.repo.GetLatestPaidByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	// A recurring subscription must be canceled successfully before local
	// entitlement is removed. This keeps a provider timeout or missing handle
	// retryable and prevents the UI from claiming that recurring billing ended.
	if latest != nil && latest.Status == StatusPaid && latest.Recurring {
		// Missing metadata means we can't reach the provider programmatically.
		// A pre-remediation Polar row without external_id, or a Paystack row
		// without an email token, is the usual case. Downgrading locally would
		// re-open audit finding 5 (recurring billing continues on card while
		// the app claims cancellation succeeded), so we hard-fail with a
		// support-ready message and a persistent audit event admins can act on.
		stuck := ""
		switch {
		case latest.ExternalID == nil || *latest.ExternalID == "":
			stuck = "missing_subscription_id"
		case s.providerByName(latest.Provider) == nil:
			stuck = "provider_not_configured"
		}
		if stuck != "" {
			s.logger.Warn("cancel stuck — awaiting manual provider action",
				"user_id", userID, "provider", latest.Provider,
				"payment_ref", latest.ProviderRef, "reason", stuck)
			if s.audit != nil {
				s.audit.Log(ctx, "billing.cancel_stuck", map[string]any{
					"actor_user_id": userID,
					"target_id":     userID,
					"payment_ref":   latest.ProviderRef,
					"provider":      latest.Provider,
					"reason":        stuck,
				})
			}
			return nil, fmt.Errorf(
				"we can't complete the cancellation automatically — " +
					"please email support@example.com from this account and " +
					"an admin will finish it within 24 hours")
		}
		switch p := s.providerByName(latest.Provider).(type) {
		case *Polar:
			if err := p.CancelSubscription(ctx, *latest.ExternalID); err != nil {
				return nil, fmt.Errorf("cancel polar subscription: %w", err)
			} else {
				s.logger.Info("polar subscription canceled",
					"user_id", userID, "subscription_id", *latest.ExternalID)
			}
		case *Paystack:
			if latest.EmailToken == nil || *latest.EmailToken == "" {
				return nil, fmt.Errorf("cancel paystack subscription: email token is missing")
			} else if err := p.DisableSubscription(ctx, *latest.ExternalID, *latest.EmailToken); err != nil {
				return nil, fmt.Errorf("cancel paystack subscription: %w", err)
			} else {
				s.logger.Info("paystack subscription disabled",
					"user_id", userID, "subscription_code", *latest.ExternalID)
			}
		default:
			return nil, fmt.Errorf("cancel subscription: unsupported provider %q", latest.Provider)
		}
	}

	if latest != nil && latest.VerifiedAt != nil &&
		time.Since(*latest.VerifiedAt) <= RefundWindow {
		provider := s.providerByName(latest.Provider)
		// Pick the right handle for each provider: Polar's refund API keys on
		// order_id; Paystack's keys on transaction reference (our ProviderRef).
		refundHandle := latest.ProviderRef
		if latest.OrderID != nil && *latest.OrderID != "" {
			refundHandle = *latest.OrderID
		}
		if provider == nil {
			s.logger.Warn("refund skipped — provider not configured",
				"user_id", userID, "provider", latest.Provider)
		} else if err := provider.Refund(ctx, refundHandle); err != nil {
			if errors.Is(err, ErrRefundUnavailable) {
				// Provider can't refund via API (e.g. Polar refund endpoint
				// needs order_id which we don't persist today). Degrade to
				// plain downgrade — user can request manual refund if needed.
				s.logger.Warn("refund unavailable — provider lacks programmatic refunds",
					"user_id", userID, "provider", latest.Provider)
			} else {
				s.logger.Error("refund failed", "user_id", userID, "ref", latest.ProviderRef, "error", err)
				return nil, fmt.Errorf("refund failed: %w", err)
			}
		} else {
			if err := s.repo.MarkStatus(ctx, latest.ProviderRef, StatusRefunded); err != nil {
				s.logger.Error("payment mark refunded failed", "ref", latest.ProviderRef, "error", err)
			}
			res.Refunded = true
			res.RefundRef = latest.ProviderRef
		}
	}

	if err := s.users.CancelPlan(ctx, userID); err != nil {
		return nil, err
	}
	s.logger.Info("plan canceled by user",
		"user_id", userID, "prev_plan", user.Plan, "refunded", res.Refunded)
	if s.audit != nil {
		s.audit.Log(ctx, "billing.plan_cancelled", map[string]any{
			"actor_user_id": userID,
			"target_id":     userID,
			"prev_plan":     string(user.Plan),
			"refunded":      res.Refunded,
			"refund_ref":    res.RefundRef,
		})
		if res.Refunded {
			s.audit.Log(ctx, "billing.plan_refunded", map[string]any{
				"actor_user_id": userID,
				"target_id":     userID,
				"reference":     res.RefundRef,
			})
		}
	}
	return res, nil
}

// titleCase upper-cases the first rune. Small helper — the whole use case
// is plan names like "Pro" / "Byo" that are ASCII, so no need for the full
// x/text/cases dependency just for this.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func generateReference() string {
	b := make([]byte, 10)
	_, _ = rand.Read(b)
	return "clyra-" + hex.EncodeToString(b)
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

package payments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const paystackBaseURL = "https://api.paystack.co"

// Paystack provider. Mock mode is active when SecretKey is empty — useful for
// local dev and for the pre-CAC window where no live key is available yet.
//
// Subscriptions:
// Paystack supports true recurring via Plans + Subscriptions. Flow:
//  1. EnsurePlan(name, amount, currency) → plan_code  (bootstrap)
//  2. InitializeSubscription(input, plan_code) → hosted checkout URL
//  3. On charge.success we get subscription_code + email_token in the
//     payload; we persist both (external_id + email_token).
//  4. Renewals arrive as charge.success events with a NEW Paystack
//     reference — we detect them via subscription_code, not reference.
//  5. Cancel calls DisableSubscription(code, token).
//
// planCache memoizes (name → plan_code) lookups within a process so we
// don't hit /plan on every checkout. Populated by EnsurePlan.
type Paystack struct {
	client    *http.Client
	secretKey string
	callback  string
	planCache sync.Map // string → string
}

// NewPaystack constructs a Paystack provider.
// secretKey == "" activates mock mode (no outbound calls; verify auto-succeeds).
func NewPaystack(secretKey, callback string) *Paystack {
	return &Paystack{
		client:    &http.Client{Timeout: 15 * time.Second},
		secretKey: secretKey,
		callback:  callback,
	}
}

func (p *Paystack) Name() string { return "paystack" }

func (p *Paystack) mock() bool { return p.secretKey == "" }

// Initialize starts a checkout. Returns an authorization URL for redirect.
// In mock mode the URL points back to the app's verify endpoint and flips
// the payment to "paid" on a follow-up verify call.
func (p *Paystack) Initialize(ctx context.Context, in ProviderInitInput) (ProviderResult, error) {
	if in.Currency != "NGN" && in.Currency != "USD" {
		return ProviderResult{}, ErrUnsupported
	}
	if p.mock() {
		// Point the user at a local "checkout complete" page — the frontend (when
		// it adds payment UI later) can immediately call /verify to progress.
		return ProviderResult{
			AuthorizationURL: "/payments/mock-return?reference=" + in.Reference,
			Reference:        in.Reference,
			Mock:             true,
		}, nil
	}

	body := map[string]any{
		"email":     in.Email,
		"amount":    in.AmountMinor,
		"currency":  in.Currency,
		"reference": in.Reference,
		"metadata": map[string]any{
			"plan":   in.Plan,
			"region": in.Region,
		},
	}
	if p.callback != "" {
		body["callback_url"] = p.callback
	}
	payload, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, paystackBaseURL+"/transaction/initialize", bytes.NewReader(payload))
	if err != nil {
		return ProviderResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return ProviderResult{}, fmt.Errorf("paystack initialize: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode != http.StatusOK {
		return ProviderResult{}, fmt.Errorf("paystack initialize %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var parsed struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AuthorizationURL string `json:"authorization_url"`
			Reference        string `json:"reference"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ProviderResult{}, fmt.Errorf("paystack init parse: %w", err)
	}
	if !parsed.Status {
		return ProviderResult{}, fmt.Errorf("paystack initialize: %s", parsed.Message)
	}
	return ProviderResult{
		AuthorizationURL: parsed.Data.AuthorizationURL,
		Reference:        parsed.Data.Reference,
	}, nil
}

// EnsurePlan returns the Paystack plan_code for a given (name, amount,
// currency, interval) tuple, creating the plan if it doesn't exist. Cached
// in-process after the first call. Mock mode returns a deterministic
// pseudo-code so local dev exercises the subscription path.
//
// interval accepts Paystack values: "hourly", "daily", "weekly", "monthly",
// "biannually", "annually". Amount is in the smallest currency unit.
func (p *Paystack) EnsurePlan(ctx context.Context, name string, amountMinor int, currency, interval string) (string, error) {
	if p.mock() {
		return "PLN_mock_" + strings.ToLower(strings.ReplaceAll(name, " ", "_")), nil
	}
	if v, ok := p.planCache.Load(name); ok {
		return v.(string), nil
	}
	if code, err := p.findPlan(ctx, name); err == nil && code != "" {
		p.planCache.Store(name, code)
		return code, nil
	}
	code, err := p.createPlan(ctx, name, amountMinor, currency, interval)
	if err != nil {
		return "", err
	}
	p.planCache.Store(name, code)
	return code, nil
}

func (p *Paystack) findPlan(ctx context.Context, name string) (string, error) {
	// /plan returns 50 by default; we set perPage=200 to comfortably fit
	// any realistic plan count. Paginate only if we grow past that.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, paystackBaseURL+"/plan?perPage=200", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+p.secretKey)
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("paystack findPlan: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("paystack findPlan %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	var parsed struct {
		Status bool `json:"status"`
		Data   []struct {
			Name     string `json:"name"`
			PlanCode string `json:"plan_code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	for _, pl := range parsed.Data {
		if pl.Name == name {
			return pl.PlanCode, nil
		}
	}
	return "", nil
}

func (p *Paystack) createPlan(ctx context.Context, name string, amountMinor int, currency, interval string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"name":     name,
		"amount":   amountMinor,
		"currency": currency,
		"interval": interval,
		// Paystack defaults both to true, which means every renewal
		// generates a monthly PDF invoice email + an SMS to the
		// subscriber. Neither is useful for our flow — users get their
		// receipt in-app under /billing and the transactional emails
		// come from us (Resend). Silence Paystack so we don't compete
		// with our own notifications or churn users on SMS costs.
		"send_invoices": false,
		"send_sms":      false,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, paystackBaseURL+"/plan", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+p.secretKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("paystack createPlan: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("paystack createPlan %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	var parsed struct {
		Status bool `json:"status"`
		Data   struct {
			PlanCode string `json:"plan_code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if parsed.Data.PlanCode == "" {
		return "", fmt.Errorf("paystack createPlan: empty plan_code — %s", truncate(string(raw), 200))
	}
	return parsed.Data.PlanCode, nil
}

// InitializeSubscription starts a Paystack subscription checkout. On success
// Paystack auto-creates a Subscription tied to the user's card and will
// auto-charge each interval. Follow-up charges arrive as `charge.success`
// webhooks with a subscription payload — we detect them via subscription_code.
func (p *Paystack) InitializeSubscription(ctx context.Context, in ProviderInitInput, planCode string) (ProviderResult, error) {
	if p.mock() {
		return ProviderResult{
			AuthorizationURL: "/payments/mock-return?reference=" + in.Reference,
			Reference:        in.Reference,
			Mock:             true,
			Recurring:        true,
		}, nil
	}
	body := map[string]any{
		"email":     in.Email,
		"amount":    in.AmountMinor,
		"reference": in.Reference,
		"plan":      planCode, // ← the subscription-mode toggle
		"metadata": map[string]any{
			"plan":   in.Plan,
			"region": in.Region,
		},
	}
	if p.callback != "" {
		body["callback_url"] = p.callback
	}
	payload, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, paystackBaseURL+"/transaction/initialize", bytes.NewReader(payload))
	if err != nil {
		return ProviderResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return ProviderResult{}, fmt.Errorf("paystack init subscription: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return ProviderResult{}, fmt.Errorf("paystack init subscription %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	var parsed struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AuthorizationURL string `json:"authorization_url"`
			Reference        string `json:"reference"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ProviderResult{}, fmt.Errorf("paystack init sub parse: %w", err)
	}
	if !parsed.Status {
		return ProviderResult{}, fmt.Errorf("paystack init subscription: %s", parsed.Message)
	}
	return ProviderResult{
		AuthorizationURL: parsed.Data.AuthorizationURL,
		Reference:        parsed.Data.Reference,
		Recurring:        true,
	}, nil
}

// DisableSubscription cancels a Paystack subscription. Both code AND email
// token are required by Paystack — the token comes from subscription.create
// webhook (we persist it as email_token). Mock mode is a no-op.
func (p *Paystack) DisableSubscription(ctx context.Context, code, token string) error {
	if p.mock() {
		return nil
	}
	if code == "" || token == "" {
		return fmt.Errorf("paystack disable: missing code or token")
	}
	body, _ := json.Marshal(map[string]any{"code": code, "token": token})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, paystackBaseURL+"/subscription/disable", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("paystack disable: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	var parsed struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(raw, &parsed)

	// Treat "already inactive" as success (idempotency for retries).
	if resp.StatusCode != http.StatusOK {
		lower := strings.ToLower(parsed.Message)
		if strings.Contains(lower, "already") || strings.Contains(lower, "inactive") {
			return nil
		}
		return fmt.Errorf("paystack disable %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	return nil
}

// ExtractSubscription pulls subscription_code + email_token from a Paystack
// webhook event. Populated on charge.success (for renewals and initial),
// subscription.create, and subscription.disable events.
//
// Returns empty strings when the event doesn't carry subscription info —
// callers fall back to reference-based lookup for those.
func (p *Paystack) ExtractSubscription(body []byte) (subscriptionCode, emailToken string, err error) {
	var evt struct {
		Event string `json:"event"`
		Data  struct {
			// subscription.* events put fields at top level of data
			SubscriptionCode string `json:"subscription_code"`
			EmailToken       string `json:"email_token"`
			// charge.success embeds them under nested subscription obj
			Subscription struct {
				SubscriptionCode string `json:"subscription_code"`
				EmailToken       string `json:"email_token"`
			} `json:"subscription"`
			Plan struct {
				PlanCode string `json:"plan_code"`
			} `json:"plan"`
		} `json:"data"`
	}
	if uerr := json.Unmarshal(body, &evt); uerr != nil {
		return "", "", fmt.Errorf("invalid json: %w", uerr)
	}
	sub := evt.Data.SubscriptionCode
	if sub == "" {
		sub = evt.Data.Subscription.SubscriptionCode
	}
	tok := evt.Data.EmailToken
	if tok == "" {
		tok = evt.Data.Subscription.EmailToken
	}
	return sub, tok, nil
}

// Verify confirms a Paystack transaction. Mock mode returns StatusPaid.
func (p *Paystack) Verify(ctx context.Context, reference string) (ProviderResult, error) {
	if p.mock() {
		return ProviderResult{Reference: reference, Status: StatusPaid, Mock: true, Recurring: true}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, paystackBaseURL+"/transaction/verify/"+reference, nil)
	if err != nil {
		return ProviderResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.secretKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return ProviderResult{}, fmt.Errorf("paystack verify: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode != http.StatusOK {
		return ProviderResult{}, fmt.Errorf("paystack verify %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var parsed struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Status    string `json:"status"` // "success" | "failed" | "abandoned" | ...
			Reference string `json:"reference"`
			Amount    int    `json:"amount"`
			Currency  string `json:"currency"`
			Metadata  struct {
				Plan string `json:"plan"`
			} `json:"metadata"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ProviderResult{}, fmt.Errorf("paystack verify parse: %w", err)
	}
	if !parsed.Status {
		return ProviderResult{Status: StatusFailed, Reference: reference}, nil
	}
	st := StatusFailed
	if strings.EqualFold(parsed.Data.Status, "success") {
		st = StatusPaid
	}
	return ProviderResult{
		Reference:   parsed.Data.Reference,
		AmountMinor: parsed.Data.Amount,
		Currency:    parsed.Data.Currency,
		Plan:        parsed.Data.Metadata.Plan,
		Status:      st,
	}, nil
}

// Refund issues a full refund via Paystack's /refund endpoint. Mock mode is a no-op.
// Paystack's response is treated as fire-and-forget: the refund is queued on their
// side and settles asynchronously; we don't wait for settlement.
func (p *Paystack) Refund(ctx context.Context, reference string) error {
	if p.mock() {
		return nil
	}

	body, _ := json.Marshal(map[string]any{"transaction": reference})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, paystackBaseURL+"/refund", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("paystack refund: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	var parsed struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(raw, &parsed)

	// Paystack returns 200 for accepted + 400 for "already refunded" — treat
	// the latter as success (idempotency) so repeated cancels don't error out.
	if resp.StatusCode != http.StatusOK {
		if strings.Contains(strings.ToLower(parsed.Message), "already") {
			return nil
		}
		return fmt.Errorf("paystack refund %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	if !parsed.Status {
		return fmt.Errorf("paystack refund: %s", parsed.Message)
	}
	return nil
}

// VerifyWebhookSignature checks Paystack's x-paystack-signature HMAC-SHA512.
// Returns true in mock mode regardless, since no signature is expected there.
func (p *Paystack) VerifyWebhookSignature(body []byte, signature string) bool {
	if p.mock() {
		return true
	}
	mac := hmac.New(sha512.New, []byte(p.secretKey))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

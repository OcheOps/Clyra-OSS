package payments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Polar is a Merchant-of-Record payment provider used for non-NG customers.
// It handles global tax compliance, so the tenant doesn't need a US entity /
// Stripe Atlas. Paired with Paystack via region-based routing in Service.
//
// Model note: Polar's checkout is *product-centric*, not amount-centric like
// Paystack. Each plan (pro / byo) maps to a Polar product UUID that holds the
// price. We pass our own `reference` as metadata so the webhook can reconcile.
type Polar struct {
	client        *http.Client
	accessToken   string
	baseURL       string
	callback      string
	webhookSecret string
	// Plan slug → Polar product UUID (set from env at bootstrap).
	productIDs map[string]string
}

// Polar base URLs. Sandbox is a real sandbox — test card, no live money.
const (
	PolarBaseProd    = "https://api.polar.sh"
	PolarBaseSandbox = "https://sandbox-api.polar.sh"
)

// NewPolar constructs a Polar provider. An empty accessToken puts the provider
// in mock mode (verify auto-succeeds, no outbound calls) — matches Paystack's
// mock behavior for local dev.
func NewPolar(accessToken, baseURL, callback, webhookSecret string, productIDs map[string]string) *Polar {
	base := baseURL
	if base == "" {
		base = PolarBaseProd
	}
	return &Polar{
		client:        &http.Client{Timeout: 15 * time.Second},
		accessToken:   accessToken,
		baseURL:       strings.TrimRight(base, "/"),
		callback:      callback,
		webhookSecret: webhookSecret,
		productIDs:    productIDs,
	}
}

func (p *Polar) Name() string { return "polar" }

func (p *Polar) mock() bool { return p.accessToken == "" }

// Initialize creates a Polar Checkout Session and returns its hosted URL.
// Our internally-minted reference is embedded in `metadata` so the webhook
// handler can reconcile back to the DB record without needing to store
// Polar's checkout ID.
func (p *Polar) Initialize(ctx context.Context, in ProviderInitInput) (ProviderResult, error) {
	if in.Currency != "USD" {
		return ProviderResult{}, ErrUnsupported
	}
	if p.mock() {
		return ProviderResult{
			AuthorizationURL: "/payments/mock-return?reference=" + in.Reference,
			Reference:        in.Reference,
			Mock:             true,
		}, nil
	}
	productID := p.productIDs[in.Plan]
	if productID == "" {
		return ProviderResult{}, fmt.Errorf("polar: no product_id configured for plan %q", in.Plan)
	}

	// Polar validates metadata values as non-empty strings (or numeric), so
	// we drop any empty field rather than sending "" — region in particular
	// is blank for global customers.
	metadata := map[string]string{}
	if in.Reference != "" {
		metadata["reference"] = in.Reference
	}
	if in.Plan != "" {
		metadata["plan"] = in.Plan
	}
	if in.Region != "" {
		metadata["region"] = in.Region
	}
	metadata["amount_minor"] = fmt.Sprint(in.AmountMinor)
	metadata["currency"] = in.Currency
	body := map[string]any{
		"products":       []string{productID},
		"customer_email": in.Email,
		"metadata":       metadata,
	}
	if p.callback != "" {
		// Polar substitutes its checkout UUID into the documented placeholder.
		// Keep our reference beside it so the backend can bind the provider
		// checkout to the exact pending payment before granting entitlement.
		sep := "?"
		if strings.Contains(p.callback, "?") {
			sep = "&"
		}
		body["success_url"] = p.callback + sep + "reference=" + url.QueryEscape(in.Reference) + "&checkout_id={CHECKOUT_ID}"
	}
	payload, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v1/checkouts/", bytes.NewReader(payload))
	if err != nil {
		return ProviderResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return ProviderResult{}, fmt.Errorf("polar initialize: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return ProviderResult{}, fmt.Errorf("polar initialize %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var parsed struct {
		ID     string `json:"id"`
		URL    string `json:"url"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ProviderResult{}, fmt.Errorf("polar init parse: %w", err)
	}
	if parsed.URL == "" {
		return ProviderResult{}, fmt.Errorf("polar: missing checkout url")
	}

	// Keep our reference as ProviderRef so GetByRef works in verify/webhook.
	// ExternalID carries Polar's UUID so the frontend can active-verify
	// (Polar does NOT append `checkout_id` to custom success_url).
	return ProviderResult{
		AuthorizationURL: parsed.URL,
		Reference:        in.Reference,
		ExternalID:       parsed.ID,
		Recurring:        true,
	}, nil
}

// Verify reconciles a Clyra payment against Polar's state using our own
// reference (echoed into Polar's metadata at checkout time). Polar supports
// metadata filtering on orders: `GET /v1/orders/?metadata=reference:<ref>`.
//
// This makes verify robust without relying on success_url query params or
// frontend sessionStorage — the backend can always find the truth on its own.
//
// Returns the subscription UUID as ExternalID so the service can persist it
// and hit Cancel/Revoke later.
func (p *Polar) Verify(ctx context.Context, reference string) (ProviderResult, error) {
	if p.mock() {
		return ProviderResult{Reference: reference, Status: StatusPaid, Mock: true}, nil
	}

	u := p.baseURL + "/v1/orders/?limit=5&metadata=" + url.QueryEscape("reference:"+reference)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return ProviderResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.accessToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return ProviderResult{}, fmt.Errorf("polar verify: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return ProviderResult{}, fmt.Errorf("polar verify %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var parsed struct {
		Items []struct {
			ID             string `json:"id"`
			Status         string `json:"status"` // "paid" | "pending" | "failed"
			Paid           bool   `json:"paid"`
			SubscriptionID string `json:"subscription_id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ProviderResult{}, fmt.Errorf("polar verify parse: %w", err)
	}
	if len(parsed.Items) == 0 {
		// No order yet — checkout probably still open. Safe pending.
		return ProviderResult{Reference: reference, Status: StatusPending}, nil
	}

	// Most recent paid order wins (Polar returns newest-first).
	for _, item := range parsed.Items {
		if item.Paid || strings.EqualFold(item.Status, "paid") {
			return ProviderResult{
				Reference:  reference,
				ExternalID: item.SubscriptionID,
				OrderID:    item.ID, // captured for the future Refund call
				Status:     StatusPaid,
			}, nil
		}
	}
	return ProviderResult{Reference: reference, Status: StatusPending}, nil
}

// LookupCheckout hits Polar's GET /v1/checkouts/{id} and maps the response
// status to our StatusPaid / StatusFailed / StatusPending. The `checkoutID`
// must be the opaque UUID Polar appends to success_url as ?checkout_id=...
// after payment.
//
// This lets the frontend reconcile a payment immediately on return without
// waiting for the webhook — critical for local dev (no public URL for Polar
// to hit) and for the "user saw success, wants plan active now" UX.
func (p *Polar) LookupCheckout(ctx context.Context, checkoutID string) (ProviderResult, error) {
	if p.mock() {
		return ProviderResult{Reference: checkoutID, Status: StatusPaid, Mock: true}, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/v1/checkouts/"+checkoutID, nil)
	if err != nil {
		return ProviderResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.accessToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return ProviderResult{}, fmt.Errorf("polar lookup checkout: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode != http.StatusOK {
		return ProviderResult{}, fmt.Errorf("polar lookup checkout %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var parsed struct {
		ID             string            `json:"id"`
		Status         string            `json:"status"` // open | expired | confirmed | succeeded | failed
		SubscriptionID string            `json:"subscription_id"`
		OrderID        string            `json:"order_id"`
		Amount         int               `json:"amount"`
		Metadata       map[string]string `json:"metadata"`
		ProductPrice   struct {
			Amount   int    `json:"price_amount"`
			Currency string `json:"price_currency"`
		} `json:"product_price"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ProviderResult{}, fmt.Errorf("polar lookup parse: %w", err)
	}

	status := StatusPending
	switch strings.ToLower(parsed.Status) {
	case "succeeded":
		status = StatusPaid
	case "failed", "expired":
		status = StatusFailed
	}
	// We return OUR reference so the service can look up the local payment row.
	// The metadata.reference round-trips through Polar intact. ExternalID
	// carries Polar's subscription UUID so we can cancel later.
	amount := parsed.Amount
	if amount <= 0 {
		amount = parsed.ProductPrice.Amount
	}
	currency := parsed.ProductPrice.Currency
	if currency == "" {
		currency = parsed.Metadata["currency"]
	}
	return ProviderResult{
		Reference:   parsed.Metadata["reference"],
		ExternalID:  parsed.SubscriptionID,
		OrderID:     parsed.OrderID,
		AmountMinor: amount,
		Currency:    currency,
		Plan:        parsed.Metadata["plan"],
		Recurring:   true,
		Status:      status,
	}, nil
}

// CancelSubscription revokes a Polar subscription immediately. `revoke=true`
// flips it to canceled + revokes access same-tick; otherwise Polar would end
// it at period_end. We use immediate revocation because our UI downgrades the
// user right away — the Polar subscription shouldn't linger.
func (p *Polar) CancelSubscription(ctx context.Context, subscriptionID string) error {
	if p.mock() {
		return nil
	}
	if subscriptionID == "" {
		return ErrRefundUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		p.baseURL+"/v1/subscriptions/"+subscriptionID+"?revoke=true", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.accessToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("polar cancel subscription: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	// 200 OK = canceled; 404 = already gone (idempotent); anything else = real error.
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}
	// 400 with "already canceled" is also idempotent in practice.
	lower := strings.ToLower(string(raw))
	if resp.StatusCode == http.StatusBadRequest && (strings.Contains(lower, "already") || strings.Contains(lower, "canceled")) {
		return nil
	}
	return fmt.Errorf("polar cancel subscription %d: %s", resp.StatusCode, truncate(string(raw), 200))
}

// Refund issues a full refund against a Polar order. The caller MUST pass
// the order_id (Payment.OrderID) — not our ProviderRef and not the
// subscription UUID. Polar's /v1/refunds/ keys on order_id only. The service
// layer pre-resolves the right handle before calling us.
//
// If the caller passes something that doesn't look like a Polar order ID
// (e.g. it forgot to pre-resolve and handed us a "clyra-xxx" reference),
// we fall back to ErrRefundUnavailable rather than burning a bad API call
// — the cancel flow degrades to "downgrade without refund" the way it used
// to, with a warning log so the gap is visible.
func (p *Polar) Refund(ctx context.Context, orderID string) error {
	if p.mock() {
		return nil
	}
	if orderID == "" || strings.HasPrefix(orderID, "clyra-") {
		return ErrRefundUnavailable
	}

	body, _ := json.Marshal(map[string]any{
		"order_id":        orderID,
		"reason":          "customer_request",
		"revoke_benefits": true,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v1/refunds/", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("polar refund: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	// 200/201 = created. 409 / 400 with "already refunded" is idempotent.
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return nil
	}
	lower := strings.ToLower(string(raw))
	if (resp.StatusCode == http.StatusConflict || resp.StatusCode == http.StatusBadRequest) &&
		strings.Contains(lower, "already") && strings.Contains(lower, "refund") {
		return nil
	}
	return fmt.Errorf("polar refund %d: %s", resp.StatusCode, truncate(string(raw), 200))
}

// VerifyWebhookSignature implements the Standard Webhooks spec that Polar uses.
// Headers: webhook-id, webhook-timestamp, webhook-signature.
// Signed content: "<id>.<timestamp>.<raw_body>"
// Signature header: "v1,<base64-sig>" (possibly multiple, space-separated).
// Secret: `whsec_<base64>` from the Polar dashboard.
func (p *Polar) VerifyWebhookSignature(body []byte, id, timestamp, signatureHeader string) bool {
	if p.mock() {
		return true
	}
	if id == "" || timestamp == "" || signatureHeader == "" || p.webhookSecret == "" {
		return false
	}
	// Polar wraps the Standard-Webhooks secret with "polar_whs_"; generic
	// Standard Webhooks hosts use "whsec_". Strip whichever is present and
	// try base64-decode, falling back to treating the remainder as raw bytes.
	secretRaw := strings.TrimPrefix(p.webhookSecret, "polar_whs_")
	secretRaw = strings.TrimPrefix(secretRaw, "whsec_")
	secret, err := base64.StdEncoding.DecodeString(secretRaw)
	if err != nil {
		secret = []byte(secretRaw)
	}

	signed := append([]byte(id+"."+timestamp+"."), body...)
	mac := hmac.New(sha256.New, secret)
	mac.Write(signed)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Reject replays older than 5 minutes.
	if ts, err := parseUnix(timestamp); err == nil {
		if time.Since(ts) > 5*time.Minute || time.Until(ts) > 5*time.Minute {
			return false
		}
	}

	for _, token := range strings.Fields(signatureHeader) {
		parts := strings.SplitN(token, ",", 2)
		if len(parts) != 2 || parts[0] != "v1" {
			continue
		}
		if hmac.Equal([]byte(parts[1]), []byte(expected)) {
			return true
		}
	}
	return false
}

// ExtractRenewal pulls the (subscription_id, order_id) pair out of a Polar
// order.paid event so the caller can distinguish a first payment (found via
// reference lookup) from a renewal (same subscription_id, new order_id, no
// reference in metadata because Polar doesn't propagate it to renewal orders).
// Returns empty strings on non-order.paid events or when either handle is
// missing — the caller falls back to reference-based lookup.
func (p *Polar) ExtractRenewal(body []byte) (subscriptionID, orderID string, err error) {
	var evt struct {
		Type string `json:"type"`
		Data struct {
			ID             string `json:"id"`
			SubscriptionID string `json:"subscription_id"`
			Subscription   struct {
				ID string `json:"id"`
			} `json:"subscription"`
		} `json:"data"`
	}
	if uerr := json.Unmarshal(body, &evt); uerr != nil {
		return "", "", fmt.Errorf("invalid json: %w", uerr)
	}
	if evt.Type != "order.paid" {
		return "", "", nil
	}
	subID := evt.Data.SubscriptionID
	if subID == "" {
		subID = evt.Data.Subscription.ID
	}
	return subID, evt.Data.ID, nil
}

// ExtractReference pulls our internal reference out of a Polar webhook body.
// Accepts either order.paid (metadata on the order) or checkout.* events.
// Returns "" + error if reference is absent so the handler can 400.
func (p *Polar) ExtractReference(body []byte) (reference, plan, region string, paid bool, err error) {
	var evt struct {
		Type string `json:"type"`
		Data struct {
			ID       string            `json:"id"`
			Status   string            `json:"status"`
			Paid     bool              `json:"paid"`
			Metadata map[string]string `json:"metadata"`
		} `json:"data"`
	}
	if uerr := json.Unmarshal(body, &evt); uerr != nil {
		return "", "", "", false, fmt.Errorf("invalid json: %w", uerr)
	}
	ref := evt.Data.Metadata["reference"]
	plan = evt.Data.Metadata["plan"]
	region = evt.Data.Metadata["region"]
	if ref == "" {
		return "", plan, region, false, errors.New("metadata.reference missing")
	}
	// Consider the payment paid when the event explicitly names it.
	switch evt.Type {
	case "order.paid":
		paid = true
	case "checkout.updated":
		paid = strings.EqualFold(evt.Data.Status, "succeeded") || evt.Data.Paid
	}
	return ref, plan, region, paid, nil
}

func parseUnix(s string) (time.Time, error) {
	// Standard Webhooks timestamp is Unix seconds as a decimal string.
	var secs int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return time.Time{}, fmt.Errorf("non-numeric timestamp")
		}
		secs = secs*10 + int64(c-'0')
	}
	return time.Unix(secs, 0), nil
}

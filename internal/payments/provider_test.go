package payments

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type paymentRoundTripFunc func(*http.Request) (*http.Response, error)

func (f paymentRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestPolarInitializeIncludesBoundCheckoutID(t *testing.T) {
	p := NewPolar("token", "https://polar.test", "https://app.test/billing/return", "secret", map[string]string{"pro": "product-1"})
	p.client = &http.Client{Transport: paymentRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		var body struct {
			SuccessURL string            `json:"success_url"`
			Metadata   map[string]string `json:"metadata"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode initialize body: %v", err)
		}
		if body.SuccessURL != "https://app.test/billing/return?reference=ref-1&checkout_id={CHECKOUT_ID}" {
			t.Fatalf("unexpected success URL %q", body.SuccessURL)
		}
		if body.Metadata["amount_minor"] != "1200" || body.Metadata["currency"] != "USD" {
			t.Fatalf("missing payment binding metadata: %#v", body.Metadata)
		}
		return jsonResponse(http.StatusCreated, `{"id":"checkout-1","url":"https://checkout.test/1"}`), nil
	})}

	result, err := p.Initialize(context.Background(), ProviderInitInput{
		Email: "person@example.test", AmountMinor: 1200, Currency: "USD",
		Plan: "pro", Reference: "ref-1",
	})
	if err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if result.ExternalID != "checkout-1" || result.Reference != "ref-1" || !result.Recurring {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestPolarLookupRequiresSucceededAndReturnsChargedAmount(t *testing.T) {
	p := NewPolar("token", "https://polar.test", "", "secret", nil)
	p.client = &http.Client{Transport: paymentRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{
			"id":"checkout-1","status":"confirmed","amount":1200,
			"subscription_id":"sub-1","order_id":"order-1",
			"metadata":{"reference":"ref-1","plan":"pro","amount_minor":"999","currency":"BAD"},
			"product_price":{"price_amount":1100,"price_currency":"USD"}
		}`), nil
	})}

	result, err := p.LookupCheckout(context.Background(), "checkout-1")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if result.Status != StatusPending {
		t.Fatalf("confirmed checkout status=%q, want pending", result.Status)
	}
	if result.AmountMinor != 1200 || result.Currency != "USD" || result.Plan != "pro" {
		t.Fatalf("provider charge fields not preserved: %+v", result)
	}
}

func TestPaystackVerifyReturnsBindingFields(t *testing.T) {
	p := NewPaystack("secret", "")
	p.client = &http.Client{Transport: paymentRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{
			"status":true,
			"data":{"status":"success","reference":"ref-1","amount":350000,"currency":"NGN","metadata":{"plan":"pro"}}
		}`), nil
	})}

	result, err := p.Verify(context.Background(), "ref-1")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if result.Status != StatusPaid || result.Reference != "ref-1" || result.AmountMinor != 350000 || result.Currency != "NGN" || result.Plan != "pro" {
		t.Fatalf("unexpected verification result: %+v", result)
	}
}

package payments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/pkg/response"
)

type Handler struct {
	service  *Service
	paystack *Paystack         // webhook signature verification (may be nil)
	polar    *Polar            // webhook signature verification (may be nil)
	idem     *IdempotencyCache // webhook replay protection (may be nil)
	audit    AuditSink         // security-event sink (may be nil)
}

func NewHandler(service *Service, paystack *Paystack, polar *Polar) *Handler {
	return &Handler{service: service, paystack: paystack, polar: polar}
}

func (h *Handler) PaystackEnabled() bool { return h.paystack != nil }
func (h *Handler) PolarEnabled() bool    { return h.polar != nil }

// WithIdempotency attaches a Redis-backed event-ID dedup cache so replayed
// webhooks are rejected before we touch the payments table.
func (h *Handler) WithIdempotency(c *IdempotencyCache) *Handler {
	h.idem = c
	return h
}

// WithAudit attaches an audit sink so signature failures + plan flips are
// surfaced to the security log.
func (h *Handler) WithAudit(a AuditSink) *Handler {
	h.audit = a
	return h
}

// AuditSink is the minimal contract the payments handler needs for audit
// logging. The concrete implementation lives in internal/audit and is
// adapted at wiring time to avoid an import cycle.
type AuditSink interface {
	Log(ctx context.Context, event string, fields map[string]any)
}

// Initialize — POST /api/v1/payments/initialize
func (h *Handler) Initialize(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req InitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}

	out, err := h.service.Initialize(r.Context(), user.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUnknownPlan), errors.Is(err, ErrNotUpgradable):
			response.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrUnsupported):
			response.Error(w, http.StatusBadRequest, "currency not supported")
		case errors.Is(err, ErrAlreadySubscribed):
			response.Error(w, http.StatusConflict, err.Error())
		case errors.Is(err, ErrTooManyPending):
			response.Error(w, http.StatusTooManyRequests, err.Error())
		default:
			slog.Error("payment initialize failed",
				"user_id", user.UserID, "plan", req.Plan, "region", req.Region, "error", err)
			response.Error(w, http.StatusInternalServerError, "payment initialization failed")
		}
		return
	}
	response.JSON(w, http.StatusOK, out)
}

// Verify — GET /api/v1/payments/verify/{reference}
func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	ref := chi.URLParam(r, "reference")
	if ref == "" {
		response.Error(w, http.StatusBadRequest, "reference required")
		return
	}

	// Optional Polar checkout_id — Polar appends it to success_url as
	// `?checkout_id=...`. When present, Service.Verify does an active lookup
	// against Polar instead of returning the "pending" placeholder.
	checkoutID := r.URL.Query().Get("checkout_id")

	rec, err := h.service.Verify(r.Context(), user.UserID, ref, VerifyHint{PolarCheckoutID: checkoutID})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "payment not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "payment verification failed")
		return
	}
	response.JSON(w, http.StatusOK, rec)
}

// List — GET /api/v1/payments
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	list, err := h.service.ListPayments(r.Context(), user.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list payments")
		return
	}
	if list == nil {
		list = []Payment{}
	}
	response.JSON(w, http.StatusOK, list)
}

// Cancel — POST /api/v1/payments/cancel
//
// Downgrades the caller to free. If their most recent payment verified within
// the refund window (24h), a full refund is queued with the provider and the
// response's `refunded` flag is true.
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	result, err := h.service.Cancel(r.Context(), user.UserID)
	if err != nil {
		if errors.Is(err, ErrNoActivePlan) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Error("plan cancel failed", "user_id", user.UserID, "error", err)
		response.Error(w, http.StatusInternalServerError, "cancel failed")
		return
	}
	response.JSON(w, http.StatusOK, result)
}

// Webhook — POST /api/v1/payments/paystack/webhook (public, signature-verified)
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "read body")
		return
	}
	if h.paystack == nil || !h.paystack.VerifyWebhookSignature(body, r.Header.Get("x-paystack-signature")) {
		if h.audit != nil {
			h.audit.Log(r.Context(), "webhook.signature_failed", map[string]any{
				"provider": "paystack", "ip": clientIP(r),
			})
		}
		response.Error(w, http.StatusUnauthorized, "invalid signature")
		return
	}

	var event struct {
		Event string `json:"event"`
		Data  struct {
			Reference string `json:"reference"`
			Status    string `json:"status"`
			ID        any    `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Dedup on Paystack's transaction id + event so retries of the same
	// charge.success don't process twice. Falls back to reference when id
	// is missing (some test events).
	dedupKey := event.Event + ":"
	switch v := event.Data.ID.(type) {
	case string:
		dedupKey += v
	case float64:
		dedupKey += jsonNumberString(v)
	default:
		dedupKey += event.Data.Reference
	}
	// Two-phase idempotency (audit finding #1). Claim gets an in-progress
	// lock with a 5m TTL; only Complete() promotes it to permanent. If
	// processing fails between the two, the 5m TTL expires and the next
	// Paystack retry gets a fresh Claim.
	switch err := h.idem.Claim(r.Context(), "paystack", dedupKey); {
	case err == nil:
		// We own the lock — proceed to process.
	case errors.Is(err, ErrAlreadyComplete):
		response.JSON(w, http.StatusOK, map[string]string{"status": "duplicate"})
		return
	case errors.Is(err, ErrProcessingInFlight):
		// Another worker (or a still-running previous attempt) has the
		// lock. Return 5xx so Paystack retries after our worker finishes
		// or the TTL expires.
		response.Error(w, http.StatusServiceUnavailable, "processing in flight, retry")
		return
	}

	subCode, emailToken, _ := h.paystack.ExtractSubscription(body)

	// Subscription lifecycle events — user or Paystack canceled the sub,
	// or Paystack failed a renewal charge. These carry a subscription_code
	// but no reference; handle them separately from charge.success and
	// short-circuit before the reference-lookup path below.
	switch event.Event {
	case "subscription.disable", "subscription.not_renew":
		if subCode != "" {
			if err := h.service.HandlePaystackSubscriptionDisabled(r.Context(), subCode); err != nil {
				slog.Error("paystack subscription disable handling failed",
					"subscription_code", subCode, "error", err)
				response.Error(w, http.StatusInternalServerError, "subscription disable failed")
				return
			}
		}
		if err := h.idem.Complete(r.Context(), "paystack", dedupKey); err != nil {
			slog.Warn("paystack idempotency complete failed",
				"dedup_key", dedupKey, "error", err)
		}
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	case "invoice.payment_failed":
		if subCode != "" {
			if err := h.service.HandlePaystackPaymentFailed(r.Context(), subCode); err != nil {
				slog.Error("paystack payment failed handling failed",
					"subscription_code", subCode, "error", err)
				response.Error(w, http.StatusInternalServerError, "payment failed handling failed")
				return
			}
		}
		if err := h.idem.Complete(r.Context(), "paystack", dedupKey); err != nil {
			slog.Warn("paystack idempotency complete failed",
				"dedup_key", dedupKey, "error", err)
		}
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}

	// Renewal path (same shape as the Polar handler): Paystack sends
	// charge.success on every recurring charge tied to a subscription. The
	// reference on renewals is Paystack-generated (not our internal one),
	// so lookup-by-reference would miss them. Dispatch via subscription_code
	// → ExternalID lookup instead.
	if event.Event == "charge.success" && subCode != "" {
		if err := h.service.HandlePaystackRenewal(r.Context(), event.Data.Reference, subCode, emailToken); err != nil {
			slog.Error("paystack renewal handling failed",
				"subscription_code", subCode, "error", err)
			// Don't Complete() — let the 5m TTL expire so Paystack's next
			// retry gets a fresh Claim + another attempt.
			response.Error(w, http.StatusInternalServerError, "renewal handling failed")
			return
		}
		// HandlePaystackRenewal is a no-op when this is the initial charge
		// (we haven't stored the subscription yet) — fall through to the
		// reference-based path so first activation still runs.
	}

	status := StatusFailed
	if event.Event == "charge.success" || event.Data.Status == "success" {
		status = StatusPaid
	}

	if err := h.service.HandleWebhook(r.Context(), "paystack", event.Data.Reference, status, WebhookMetadata{
		ExternalID: subCode,
		EmailToken: emailToken,
	}); err != nil && !errors.Is(err, ErrNotFound) {
		// Same treatment — leave the lock in-progress so it TTLs out
		// and Paystack's retry gets to try again.
		response.Error(w, http.StatusInternalServerError, "webhook handling failed")
		return
	}

	// Only NOW do we promote the idempotency lock. If Complete fails
	// (Redis down), the worst case is a duplicate reprocess after 5m —
	// the service layer's StatusPaid short-circuit prevents double-upgrade.
	if err := h.idem.Complete(r.Context(), "paystack", dedupKey); err != nil {
		slog.Warn("paystack idempotency complete failed",
			"dedup_key", dedupKey, "error", err)
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// PolarWebhook — POST /api/v1/payments/polar/webhook (public, signature-verified).
// Polar uses Standard Webhooks: webhook-id / webhook-timestamp / webhook-signature.
func (h *Handler) PolarWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "read body")
		return
	}
	if h.polar == nil {
		response.Error(w, http.StatusNotImplemented, "polar not configured")
		return
	}
	id := r.Header.Get("webhook-id")
	ts := r.Header.Get("webhook-timestamp")
	sig := r.Header.Get("webhook-signature")
	if id == "" || !h.polar.VerifyWebhookSignature(body, id, ts, sig) {
		if h.audit != nil {
			h.audit.Log(r.Context(), "webhook.signature_failed", map[string]any{
				"provider": "polar", "webhook_id": id, "ip": clientIP(r),
			})
		}
		response.Error(w, http.StatusUnauthorized, "invalid signature")
		return
	}

	// Polar webhook-id is unique per delivery — perfect dedup key.
	// Two-phase idempotency (audit finding #1) — see paystack handler.
	switch err := h.idem.Claim(r.Context(), "polar", id); {
	case err == nil:
	case errors.Is(err, ErrAlreadyComplete):
		response.JSON(w, http.StatusOK, map[string]string{"status": "duplicate"})
		return
	case errors.Is(err, ErrProcessingInFlight):
		response.Error(w, http.StatusServiceUnavailable, "processing in flight, retry")
		return
	}

	// Renewal path first: an order.paid from a recurring subscription
	// carries subscription_id + order_id but NOT our metadata.reference
	// (Polar only propagates checkout metadata to the first order). Skipping
	// this used to drop every renewal silently.
	subID, orderID, rerr := h.polar.ExtractRenewal(body)
	if rerr == nil && subID != "" && orderID != "" {
		if err := h.service.HandleRenewal(r.Context(), subID, orderID); err != nil {
			slog.Error("polar renewal handling failed",
				"subscription_id", subID, "order_id", orderID, "error", err)
			// Don't Complete — TTL out so Polar retries.
			response.Error(w, http.StatusInternalServerError, "renewal handling failed")
			return
		}
	}

	reference, _, _, paid, perr := h.polar.ExtractReference(body)
	if perr != nil {
		// Renewals don't carry metadata.reference — if we already processed
		// the renewal above, ignore the missing reference and 200.
		if subID != "" && orderID != "" {
			if err := h.idem.Complete(r.Context(), "polar", id); err != nil {
				slog.Warn("polar idempotency complete failed", "webhook_id", id, "error", err)
			}
			response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}
		// Genuinely bad payload — release the lock so a subsequent valid
		// delivery of the same webhook-id (unlikely but possible) doesn't
		// wait 5m.
		_ = h.idem.Release(r.Context(), "polar", id)
		response.Error(w, http.StatusBadRequest, perr.Error())
		return
	}

	status := StatusPending
	if paid {
		status = StatusPaid
	}
	if err := h.service.HandleWebhook(r.Context(), "polar", reference, status, WebhookMetadata{
		ExternalID: subID,
		OrderID:    orderID,
	}); err != nil && !errors.Is(err, ErrNotFound) {
		slog.Error("polar webhook handling failed", "ref", reference, "error", err)
		response.Error(w, http.StatusInternalServerError, "webhook handling failed")
		return
	}

	if err := h.idem.Complete(r.Context(), "polar", id); err != nil {
		slog.Warn("polar idempotency complete failed", "webhook_id", id, "error", err)
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- helpers ---

func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		// First entry is the originating client; the rest are proxies.
		for i := 0; i < len(v); i++ {
			if v[i] == ',' {
				return v[:i]
			}
		}
		return v
	}
	if v := r.Header.Get("X-Real-IP"); v != "" {
		return v
	}
	return r.RemoteAddr
}

func jsonNumberString(f float64) string {
	// Paystack ids fit in int64; format without scientific notation.
	if f == float64(int64(f)) {
		return strconvI(int64(f))
	}
	return ""
}

func strconvI(i int64) string {
	// avoid importing strconv just for this — fmt.Sprint is fine here.
	return fmt.Sprint(i)
}

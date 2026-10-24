// Package audit is the append-only security event log. It collects events
// that auditors and incident responders need: plan changes, role changes,
// share-token mints/revokes, IaC exports, login + password reset, and
// webhook signature failures.
//
// Design notes:
//   - Writes go to the `audit_events` table (see schema.sql). We never
//     UPDATE or DELETE — production grants should reflect this.
//   - Failures to write are logged but never bubble up to the caller.
//     Audit logging is best-effort: a transient DB hiccup must not block
//     a paying user from cancelling, or stop a webhook ack.
//   - The Sink interface lets unit tests substitute an in-memory recorder
//     and lets handlers in other packages depend on a tiny surface
//     (avoiding import cycles).
package audit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Event names. Keep these as constants so callers don't free-form spell them
// inconsistently across packages.
const (
	EventLoginSuccess          = "auth.login_success"
	EventPasswordResetRequest  = "auth.password_reset_requested"
	EventPasswordResetComplete = "auth.password_reset_completed"
	EventPlanUpgraded          = "billing.plan_upgraded"
	EventPlanCancelled         = "billing.plan_cancelled"
	EventPlanRefunded          = "billing.plan_refunded"
	EventRoleChanged           = "admin.role_changed"
	EventShareCreated          = "share.created"
	EventShareRevoked          = "share.revoked"
	EventIaCExport             = "iac.exported"
	EventWebhookSigFailed      = "webhook.signature_failed"
	EventMCPToolCalled         = "mcp.tool_called"
)

// Sink is the minimal interface consumers depend on. The concrete Recorder
// satisfies it. Tests can swap in a recording fake.
type Sink interface {
	Log(ctx context.Context, event string, fields map[string]any)
}

// Recorder writes audit events to Postgres.
type Recorder struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewRecorder(pool *pgxpool.Pool, logger *slog.Logger) *Recorder {
	return &Recorder{pool: pool, logger: logger}
}

// Log persists one audit event. Recognized fields:
//
//	"actor_user_id" string  — who did it (nil for system / webhook actors)
//	"target_id"     string  — primary subject id (e.g. project_id, user_id of target)
//	"ip"            string  — request origin
//	"user_agent"    string  — request user-agent
//
// Anything else is folded into the metadata JSONB column.
func (r *Recorder) Log(ctx context.Context, event string, fields map[string]any) {
	if r == nil || r.pool == nil || event == "" {
		return
	}

	id, err := newID()
	if err != nil {
		r.warn("audit: id generation failed", "error", err)
		return
	}

	var actor, target, ip, ua *string
	meta := map[string]any{}
	for k, v := range fields {
		switch k {
		case "actor_user_id":
			actor = strPtr(v)
		case "target_id":
			target = strPtr(v)
		case "ip":
			ip = strPtr(v)
		case "user_agent":
			ua = strPtr(v)
		default:
			meta[k] = v
		}
	}

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		// Bad metadata shouldn't block the audit row — fall back to empty obj.
		metaJSON = []byte(`{}`)
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO audit_events (id, event, actor_user_id, target_id, ip, user_agent, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, event, actor, target, ip, ua, metaJSON,
	)
	if err != nil {
		// Failing to record an audit event is bad but must not break the
		// caller. Surface to the operational log so SREs notice.
		r.warn("audit: insert failed", "event", event, "error", err)
	}
}

func (r *Recorder) warn(msg string, args ...any) {
	if r.logger != nil {
		r.logger.Warn(msg, args...)
	}
}

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "aud_" + hex.EncodeToString(b), nil
}

func strPtr(v any) *string {
	s, ok := v.(string)
	if !ok || s == "" {
		return nil
	}
	return &s
}

// Event is the read shape used by the admin viewer.
type Event struct {
	ID          string         `json:"id"`
	OccurredAt  time.Time      `json:"occurred_at"`
	Event       string         `json:"event"`
	ActorUserID *string        `json:"actor_user_id,omitempty"`
	TargetID    *string        `json:"target_id,omitempty"`
	IP          *string        `json:"ip,omitempty"`
	UserAgent   *string        `json:"user_agent,omitempty"`
	Metadata    map[string]any `json:"metadata"`
}

// ListFilter narrows the result set. Empty fields are ignored.
type ListFilter struct {
	Event       string
	ActorUserID string
	Limit       int
	Offset      int
}

// List returns the most-recent events matching filter, newest first. Hard
// upper bound of 500 per page so a too-greedy admin click doesn't pin the
// API process. Total is the count matching the filter (ignoring paging).
func (r *Recorder) List(ctx context.Context, f ListFilter) ([]Event, int, error) {
	if r == nil || r.pool == nil {
		return nil, 0, nil
	}
	limit := f.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	// Build the WHERE clause dynamically but with positional placeholders —
	// no string concat into the SQL itself.
	where := "WHERE 1=1"
	args := []any{}
	if f.Event != "" {
		args = append(args, f.Event)
		where += " AND event = $" + itoa(len(args))
	}
	if f.ActorUserID != "" {
		args = append(args, f.ActorUserID)
		where += " AND actor_user_id = $" + itoa(len(args))
	}

	// Total count (small table for now; revisit if it grows past ~1M rows).
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_events "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	q := "SELECT id, occurred_at, event, actor_user_id, target_id, ip, user_agent, metadata " +
		"FROM audit_events " + where +
		" ORDER BY occurred_at DESC LIMIT $" + itoa(len(args)-1) + " OFFSET $" + itoa(len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]Event, 0, limit)
	for rows.Next() {
		var e Event
		var metaRaw []byte
		if err := rows.Scan(&e.ID, &e.OccurredAt, &e.Event, &e.ActorUserID, &e.TargetID, &e.IP, &e.UserAgent, &metaRaw); err != nil {
			return nil, 0, err
		}
		if len(metaRaw) > 0 {
			_ = json.Unmarshal(metaRaw, &e.Metadata)
		}
		if e.Metadata == nil {
			e.Metadata = map[string]any{}
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

func itoa(n int) string {
	// Tiny helper to avoid importing strconv just for placeholder indexing.
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

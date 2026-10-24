// Package admin owns the admin-only HTTP surface for user management
// (list, role, plan grant, metrics, single email). Gated behind the
// existing RequireAdmin middleware — this package does not re-check
// roles.
package admin

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oche/diagram/internal/audit"
	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/internal/repository"
	"github.com/oche/diagram/pkg/response"
)

// UserStore is the subset of the user repo this handler needs.
type UserStore interface {
	List(ctx context.Context, query string, limit, offset int) ([]repository.UserListRow, int, error)
	GetByID(ctx context.Context, id string) (*auth.User, error)
	UpdateRole(ctx context.Context, userID string, role auth.Role) error
	UpgradePlan(ctx context.Context, userID string, plan auth.Plan, expiry time.Time) error
	UpdatePlan(ctx context.Context, userID string, plan auth.Plan) error
	CancelPlan(ctx context.Context, userID string) error
	CountAdmins(ctx context.Context) (int, error)
}

// EmailSender is the one method we need from invite.EmailService. Declared
// as an interface so tests can swap it.
type EmailSender interface {
	SendAdmin(toEmail, subject, body string) error
	Enabled() bool
}

// AuditSink mirrors the contract in other packages so admin can emit
// role/plan-change events without an import cycle.
type AuditSink interface {
	Log(ctx context.Context, event string, fields map[string]any)
}

type Handler struct {
	users    UserStore
	email    EmailSender
	pool     *pgxpool.Pool
	audit    AuditSink
	auditR   *audit.Recorder // read side for the viewer endpoint
	logger   *slog.Logger
}

func NewHandler(users UserStore, email EmailSender, pool *pgxpool.Pool, logger *slog.Logger) *Handler {
	return &Handler{users: users, email: email, pool: pool, logger: logger}
}

// WithAudit attaches the audit recorder. Optional. The same value is used
// both as a write-side sink (Log) and a read-side source for the viewer.
func (h *Handler) WithAudit(a *audit.Recorder) *Handler {
	h.audit = a
	h.auditR = a
	return h
}

// ListAudit — GET /api/v1/admin/audit?event=&actor=&limit=&offset=
//
// Returns audit_events newest-first. Filter params are optional. Limit is
// hard-capped at 500 by the recorder; default 100. The response shape is
// {items: [...], total: N} so the UI can paginate.
func (h *Handler) ListAudit(w http.ResponseWriter, r *http.Request) {
	if h.auditR == nil {
		response.Error(w, http.StatusServiceUnavailable, "audit log unavailable")
		return
	}
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	items, total, err := h.auditR.List(r.Context(), audit.ListFilter{
		Event:       q.Get("event"),
		ActorUserID: q.Get("actor"),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		h.logger.Error("audit list failed", "error", err)
		response.Error(w, http.StatusInternalServerError, "audit query failed")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

// ─── Users: list / role / plan ────────────────────────────────────────

type userRow struct {
	ID         string     `json:"id"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Plan       string     `json:"plan"`
	PlanExpiry *time.Time `json:"plan_expiry,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type listUsersResponse struct {
	Users []userRow `json:"users"`
	Total int       `json:"total"`
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	rows, total, err := h.users.List(r.Context(), q, limit, offset)
	if err != nil {
		h.logger.Error("admin list users", "error", err)
		response.Error(w, http.StatusInternalServerError, "list failed")
		return
	}
	out := make([]userRow, 0, len(rows))
	for _, u := range rows {
		out = append(out, userRow{
			ID: u.ID, Email: u.Email, Role: u.Role, Plan: u.Plan,
			PlanExpiry: u.PlanExpiry, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
		})
	}
	response.JSON(w, http.StatusOK, listUsersResponse{Users: out, Total: total})
}

type setRoleRequest struct {
	Role string `json:"role"`
}

func (h *Handler) SetRole(w http.ResponseWriter, r *http.Request) {
	actor := auth.UserFromContext(r.Context())
	userID := pathTail(r.URL.Path, "/role")
	if userID == "" {
		response.Error(w, http.StatusBadRequest, "user id required")
		return
	}
	var req setRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	req.Role = strings.ToLower(strings.TrimSpace(req.Role))
	if req.Role != "admin" && req.Role != "user" {
		response.Error(w, http.StatusBadRequest, "role must be admin or user")
		return
	}

	// Safety: can't demote the last admin. Also can't demote yourself —
	// leaves you with a single click to lock the product.
	if req.Role == "user" {
		if actor != nil && actor.UserID == userID {
			response.Error(w, http.StatusBadRequest, "you cannot demote yourself")
			return
		}
		n, err := h.users.CountAdmins(r.Context())
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "role check failed")
			return
		}
		target, err := h.users.GetByID(r.Context(), userID)
		if err != nil {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		if target.Role == auth.RoleAdmin && n <= 1 {
			response.Error(w, http.StatusBadRequest, "cannot demote the last admin")
			return
		}
	}

	if err := h.users.UpdateRole(r.Context(), userID, auth.Role(req.Role)); err != nil {
		h.logger.Error("admin set role", "user_id", userID, "error", err)
		response.Error(w, http.StatusInternalServerError, "update failed")
		return
	}
	if actor != nil {
		h.logger.Info("admin changed role", "actor", actor.UserID, "target", userID, "role", req.Role)
	}
	if h.audit != nil && actor != nil {
		h.audit.Log(r.Context(), "admin.role_changed", map[string]any{
			"actor_user_id": actor.UserID,
			"target_id":     userID,
			"new_role":      req.Role,
			"ip":            clientIP(r),
		})
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type grantPlanRequest struct {
	Plan      string     `json:"plan"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"` // nil = never (no expiry)
	Note      string     `json:"note,omitempty"`
}

// GrantPlan sets the user's plan and expiry, and records a `payments` row
// tagged as a grant so the billing history UI shows it alongside real
// payments. Amount is zero; currency inherited from the user's locale is
// skipped — we just set NGN to match the existing schema non-null column.
func (h *Handler) GrantPlan(w http.ResponseWriter, r *http.Request) {
	actor := auth.UserFromContext(r.Context())
	userID := pathTail(r.URL.Path, "/plan")
	if userID == "" {
		response.Error(w, http.StatusBadRequest, "user id required")
		return
	}

	var req grantPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	plan := strings.ToLower(strings.TrimSpace(req.Plan))
	if plan != "free" && plan != "pro" && plan != "byo" {
		response.Error(w, http.StatusBadRequest, "plan must be free, pro, or byo")
		return
	}

	if plan == "free" {
		if err := h.users.CancelPlan(r.Context(), userID); err != nil {
			response.Error(w, http.StatusInternalServerError, "cancel failed")
			return
		}
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}

	// Default expiry: 30 days if not provided.
	expiry := time.Now().UTC().Add(30 * 24 * time.Hour)
	if req.ExpiresAt != nil {
		expiry = req.ExpiresAt.UTC()
	}
	if err := h.users.UpgradePlan(r.Context(), userID, auth.Plan(plan), expiry); err != nil {
		response.Error(w, http.StatusInternalServerError, "grant failed")
		return
	}

	// Record the grant in payments so it shows up in the user's billing
	// history. Best-effort — the plan change above is what matters.
	ref := "grant-" + randRef()
	_, grantErr := h.pool.Exec(r.Context(),
		`INSERT INTO payments
		   (id, user_id, provider, provider_ref, amount_minor, currency, region, plan, period_days, status, created_at, verified_at)
		 VALUES ($1, $2, 'grant', $3, 0, 'NGN', '', $4, $5, 'paid', NOW(), NOW())
		 ON CONFLICT (provider_ref) DO NOTHING`,
		ref, userID, ref, plan, int(time.Until(expiry).Hours()/24),
	)
	if grantErr != nil {
		h.logger.Warn("admin grant payment row failed", "error", grantErr)
	}

	if actor != nil {
		h.logger.Info("admin granted plan",
			"actor", actor.UserID, "target", userID,
			"plan", plan, "expiry", expiry, "note", req.Note,
		)
	}
	if h.audit != nil && actor != nil {
		h.audit.Log(r.Context(), "billing.plan_upgraded", map[string]any{
			"actor_user_id": actor.UserID,
			"target_id":     userID,
			"plan":          plan,
			"source":        "admin_grant",
			"note":          req.Note,
			"ip":            clientIP(r),
		})
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ─── Email a user ─────────────────────────────────────────────────────

type emailRequest struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func (h *Handler) EmailUser(w http.ResponseWriter, r *http.Request) {
	userID := pathTail(r.URL.Path, "/email")
	if userID == "" {
		response.Error(w, http.StatusBadRequest, "user id required")
		return
	}
	target, err := h.users.GetByID(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}
	var req emailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	req.Subject = strings.TrimSpace(req.Subject)
	req.Body = strings.TrimSpace(req.Body)
	if req.Subject == "" || req.Body == "" {
		response.Error(w, http.StatusBadRequest, "subject and body required")
		return
	}
	if len(req.Subject) > 160 {
		response.Error(w, http.StatusBadRequest, "subject too long")
		return
	}
	if len(req.Body) > 8000 {
		response.Error(w, http.StatusBadRequest, "body too long (8k max)")
		return
	}
	if !h.email.Enabled() {
		response.Error(w, http.StatusServiceUnavailable, "email service not configured")
		return
	}
	if err := h.email.SendAdmin(target.Email, req.Subject, req.Body); err != nil {
		h.logger.Error("admin email failed", "target", userID, "error", err)
		response.Error(w, http.StatusBadGateway, "send failed")
		return
	}
	actor := auth.UserFromContext(r.Context())
	if actor != nil {
		h.logger.Info("admin emailed user", "actor", actor.UserID, "target", userID, "subject", req.Subject)
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ─── Metrics ──────────────────────────────────────────────────────────

type metrics struct {
	TotalUsers         int `json:"total_users"`
	UsersByPlan        map[string]int `json:"users_by_plan"`
	SignupsLast7d      int `json:"signups_last_7d"`
	SignupsLast30d     int `json:"signups_last_30d"`
	ActivationFunnel   map[string]int `json:"activation_funnel"`
}

func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	m := metrics{
		UsersByPlan:      map[string]int{},
		ActivationFunnel: map[string]int{},
	}

	if err := h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&m.TotalUsers); err != nil {
		h.logger.Error("admin metrics: total_users", "error", err)
	}

	if rows, err := h.pool.Query(ctx, `SELECT plan, COUNT(*) FROM users GROUP BY plan`); err == nil {
		defer rows.Close()
		for rows.Next() {
			var k string
			var v int
			if err := rows.Scan(&k, &v); err == nil {
				m.UsersByPlan[k] = v
			}
		}
	}

	_ = h.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE created_at > NOW() - INTERVAL '7 days'`).
		Scan(&m.SignupsLast7d)
	_ = h.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE created_at > NOW() - INTERVAL '30 days'`).
		Scan(&m.SignupsLast30d)

	if rows, err := h.pool.Query(ctx, `SELECT kind, COUNT(*) FROM activation_events GROUP BY kind`); err == nil {
		defer rows.Close()
		for rows.Next() {
			var k string
			var v int
			if err := rows.Scan(&k, &v); err == nil {
				m.ActivationFunnel[k] = v
			}
		}
	}

	response.JSON(w, http.StatusOK, m)
}

// ─── Utilities ────────────────────────────────────────────────────────

// pathTail pulls the user id out of paths like
// /api/v1/admin/users/{id}/role → {id}. Chi's URL params would be
// cleaner, but these routes are registered as literal paths in the
// existing router for consistency with the rest of the admin group.
func pathTail(path, suffix string) string {
	path = strings.TrimSuffix(path, suffix)
	// strip `/api/v1/admin/users/` prefix
	idx := strings.LastIndex(path, "/users/")
	if idx == -1 {
		return ""
	}
	return strings.Trim(path[idx+len("/users/"):], "/")
}

// randRef / grantMetadata are small helpers that avoid pulling a new
// dependency just for random IDs + JSON marshalling.
func randRef() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// Keep errors imported in case future callers need custom sentinels.
var _ = errors.New

func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
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

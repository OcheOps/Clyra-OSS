package invite

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/pkg/response"
)

type Handler struct {
	repo  *Repository
	email *EmailService
}

func NewHandler(repo *Repository, email *EmailService) *Handler {
	return &Handler{repo: repo, email: email}
}

// ValidateCode checks if an invite code is valid (public — no auth needed).
func (h *Handler) ValidateCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		response.Error(w, http.StatusBadRequest, "invite code is required")
		return
	}

	inv, err := h.repo.GetByCode(r.Context(), req.Code)
	if err != nil || inv == nil || !inv.IsValid() {
		response.Error(w, http.StatusUnprocessableEntity,
			"This invite code isn't valid or has already been used.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"valid": true})
}

// CreateInvite generates a new invite code and optionally sends an email (admin only).
func (h *Handler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Email string `json:"email"`
		Plan  string `json:"plan"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	plan := normalizePlan(req.Plan)

	inv, err := h.repo.Create(r.Context(), user.UserID, nil, plan)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create invite")
		return
	}

	// Send email if provided and email service is configured
	emailSent := false
	if req.Email != "" && h.email != nil && h.email.Enabled() {
		if err := h.email.SendInvite(req.Email, inv.Code); err == nil {
			emailSent = true
		}
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"invite":     inv,
		"email_sent": emailSent,
		"email_to":   req.Email,
	})
}

// ListInvites returns all invite codes (admin only).
func (h *Handler) ListInvites(w http.ResponseWriter, r *http.Request) {
	invites, err := h.repo.ListAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list invites")
		return
	}
	if invites == nil {
		invites = []InviteCode{}
	}

	response.JSON(w, http.StatusOK, invites)
}

// RequestAccess is the public endpoint for "I don't have an invite" submissions.
func (h *Handler) RequestAccess(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Note  string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.Note = strings.TrimSpace(req.Note)
	if _, err := mail.ParseAddress(req.Email); err != nil {
		response.Error(w, http.StatusBadRequest, "a valid email is required")
		return
	}
	if len(req.Note) > 500 {
		req.Note = req.Note[:500]
	}

	// One pending request per email per hour — keeps the inbox tidy and
	// blocks naive scripted submissions.
	recent, err := h.repo.HasRecentPendingRequest(r.Context(), req.Email, time.Hour)
	if err == nil && recent {
		// Respond 202 to avoid leaking whether an email is already queued —
		// same shape as a fresh submission.
		response.JSON(w, http.StatusAccepted, map[string]any{"status": "received"})
		return
	}

	rec, err := h.repo.CreateRequest(r.Context(), req.Email, req.Note)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to record request")
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{"status": "received", "id": rec.ID})
}

// ListRequests returns all invite requests (admin only).
func (h *Handler) ListRequests(w http.ResponseWriter, r *http.Request) {
	status := RequestStatus(r.URL.Query().Get("status"))
	reqs, err := h.repo.ListRequests(r.Context(), status)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list requests")
		return
	}
	if reqs == nil {
		reqs = []Request{}
	}
	response.JSON(w, http.StatusOK, reqs)
}

// ApproveRequest converts a pending request into an invite code (admin only).
func (h *Handler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "request id is required")
		return
	}

	var body struct {
		Plan string `json:"plan"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	plan := normalizePlan(body.Plan)

	rec, err := h.repo.GetRequest(r.Context(), id)
	if err != nil || rec == nil {
		response.Error(w, http.StatusNotFound, "request not found")
		return
	}
	if rec.Status != StatusPending {
		response.Error(w, http.StatusConflict, "request already handled")
		return
	}

	inv, err := h.repo.Create(r.Context(), user.UserID, nil, plan)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create invite")
		return
	}

	if err := h.repo.MarkRequestHandled(r.Context(), id, StatusApproved, user.UserID); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update request")
		return
	}

	emailSent := false
	if h.email != nil && h.email.Enabled() {
		if err := h.email.SendInvite(rec.Email, inv.Code); err == nil {
			emailSent = true
		}
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"invite":     inv,
		"email_sent": emailSent,
		"email_to":   rec.Email,
	})
}

// DeclineRequest marks a pending request as declined (admin only).
func (h *Handler) DeclineRequest(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "request id is required")
		return
	}

	if err := h.repo.MarkRequestHandled(r.Context(), id, StatusDeclined, user.UserID); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update request")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "declined"})
}

// DeleteInvite removes an invite code (admin only).
func (h *Handler) DeleteInvite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		response.Error(w, http.StatusBadRequest, "invite id is required")
		return
	}

	if err := h.repo.Delete(r.Context(), req.ID); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete invite")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

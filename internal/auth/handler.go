package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/oche/diagram/pkg/validate"
	"github.com/oche/diagram/pkg/response"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service, logger: slog.Default()}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if !validate.DecodeAndValidate(w, r, &req) {
		return
	}

	user, cookie, err := h.service.Register(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	http.SetCookie(w, cookie)
	http.SetCookie(w, h.service.IssueCSRFCookie())
	response.JSON(w, http.StatusCreated, AuthResponse{
		User: &UserProfile{ID: user.ID, Email: user.Email, Role: user.Role, Plan: user.Plan, PlanExpiry: user.PlanExpiry, EffectivePlan: user.EffectivePlan()},
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !validate.DecodeAndValidate(w, r, &req) {
		return
	}

	user, cookie, err := h.service.Login(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	http.SetCookie(w, cookie)
	http.SetCookie(w, h.service.IssueCSRFCookie())
	response.JSON(w, http.StatusOK, AuthResponse{
		User: &UserProfile{ID: user.ID, Email: user.Email, Role: user.Role, Plan: user.Plan, PlanExpiry: user.PlanExpiry, EffectivePlan: user.EffectivePlan()},
	})
}

// Me — GET /api/v1/auth/me. Returns the authenticated user's profile so the
// SPA can hydrate its auth store on mount (post-OAuth redirect, page reload,
// fresh tab). Authenticate middleware has already validated the session.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	su := UserFromContext(r.Context())
	if su == nil {
		// Diagnostic breadcrumb — distinguishes "no cookie at all" (browser
		// never received Set-Cookie, e.g. SameSite/Secure/Domain mismatch)
		// from "cookie present but invalid" (signing mismatch, expired
		// session, DB row gone). Triggered only on the 401 path so this
		// stays low-volume.
		sessionCookie := h.service.cfg.Session.CookieName
		_, hasSession := r.Cookie(sessionCookie)
		_, hasCSRF := r.Cookie(sessionCookie + "_csrf")
		h.logger.Warn("auth/me unauthenticated",
			"has_session_cookie", hasSession,
			"has_csrf_cookie", hasCSRF,
			"user_agent", r.UserAgent(),
			"referer", r.Referer(),
		)
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	response.JSON(w, http.StatusOK, AuthResponse{
		User: &UserProfile{
			ID:            su.UserID,
			Email:         su.Email,
			Role:          su.Role,
			Plan:          su.Plan,
			PlanExpiry:    su.PlanExpiry,
			EffectivePlan: su.EffectivePlan(),
		},
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	for _, c := range h.service.Logout(r.Context(), r) {
		http.SetCookie(w, c)
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// ForgotPassword — POST /api/v1/auth/forgot-password (public).
//
// Always responds 200 regardless of whether the email exists. This prevents
// account enumeration while still sending a reset link to real users.
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if !validate.DecodeAndValidate(w, r, &req) {
		return
	}
	// Service swallows "unknown email" internally — we never return an error.
	_ = h.service.RequestPasswordReset(r.Context(), req.Email)
	response.JSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "If that email is registered, a reset link is on its way.",
	})
}

// ResetPassword — POST /api/v1/auth/reset-password (public).
// Consumes a single-use token from the email link.
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if !validate.DecodeAndValidate(w, r, &req) {
		return
	}
	if err := h.service.ResetPassword(r.Context(), req.Token, req.Password); err != nil {
		h.handleError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrAccountLocked):
		response.Error(w, http.StatusTooManyRequests, err.Error())
	case errors.Is(err, ErrInvalidCredentials):
		response.Error(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrEmailTaken):
		response.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrInvalidInvite):
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrWeakPassword):
		response.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrInvalidSession):
		response.Error(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrResetTokenInvalid), errors.Is(err, ErrResetTokenExpired):
		response.Error(w, http.StatusBadRequest, err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

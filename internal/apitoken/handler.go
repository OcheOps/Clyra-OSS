package apitoken

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/pkg/response"
)

// Handler exposes the /settings/mcp/tokens endpoints. Session-authenticated
// (SPA users only) — token-authenticated callers can't manage other tokens.
type Handler struct {
	svc    *Service
	logger *slog.Logger
}

func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{svc: svc, logger: logger}
}

type createRequest struct {
	Name string `json:"name"`
}

type listResponse struct {
	Tokens []Token `json:"tokens"`
}

type createResponse struct {
	Token Token `json:"token"`
}

// Create — POST /api/v1/settings/mcp/tokens
// Mints a token. Response includes the cleartext ONCE via Token.Clear.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	su := auth.UserFromContext(r.Context())
	if su == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	t, err := h.svc.Mint(r.Context(), su.UserID, req.Name)
	switch {
	case errors.Is(err, ErrInvalidName):
		response.Error(w, http.StatusBadRequest, "name required (max 64 chars)")
		return
	case errors.Is(err, ErrTokenLimit):
		response.Error(w, http.StatusUnprocessableEntity, "token limit reached — revoke an existing token first")
		return
	case err != nil:
		h.logger.Error("mint api token failed", "user_id", su.UserID, "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to create token")
		return
	}
	response.JSON(w, http.StatusCreated, createResponse{Token: *t})
}

// List — GET /api/v1/settings/mcp/tokens
// Returns all of the caller's tokens (live + revoked). No cleartext.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	su := auth.UserFromContext(r.Context())
	if su == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tokens, err := h.svc.List(r.Context(), su.UserID)
	if err != nil {
		h.logger.Error("list api tokens failed", "user_id", su.UserID, "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to list tokens")
		return
	}
	if tokens == nil {
		tokens = []Token{}
	}
	response.JSON(w, http.StatusOK, listResponse{Tokens: tokens})
}

// Revoke — DELETE /api/v1/settings/mcp/tokens/{id}
// Marks the token revoked. Idempotent — a second call is a no-op 204.
func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	su := auth.UserFromContext(r.Context())
	if su == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "token id required")
		return
	}
	if err := h.svc.Revoke(r.Context(), su.UserID, id); err != nil {
		h.logger.Error("revoke api token failed",
			"user_id", su.UserID, "token_id", id, "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to revoke token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

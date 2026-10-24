package iac

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/oche/diagram/internal/abuse"
	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/internal/provider"
	"github.com/oche/diagram/pkg/response"
	"github.com/oche/diagram/pkg/validate"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service, logger: slog.Default()}
}

// Generate handles POST /api/v1/iac/generate.
func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req GenerateRequest
	if !validate.DecodeAndValidate(w, r, &req) {
		return
	}

	// Block non-BYO users from passing their own provider key — the frontend
	// hides BYO providers behind the BYO plan but this endpoint is reachable
	// directly. Without this a free user could curl `api_key=...` and burn
	// our managed quota in their name. Mirrors the diagram handler.
	if user.EffectivePlan() != auth.PlanBYO && req.APIKey != "" {
		response.Error(w, http.StatusForbidden, "BYO keys require the BYO plan")
		return
	}

	result, err := h.service.Generate(r.Context(), user.UserID, string(user.EffectivePlan()), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrModeNotSupported),
		errors.Is(err, ErrDiagramEmpty),
		errors.Is(err, ErrDiagramTooLarge),
		errors.Is(err, ErrFormatUnsupported),
		errors.Is(err, ErrCloudUnsupported),
		errors.Is(err, ErrMissingProvider),
		errors.Is(err, ErrMissingAPIKey):
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrFreePlanIaCUsed) {
		response.Error(w, http.StatusPaymentRequired, err.Error())
		return
	}

	// Abuse limits
	if errors.Is(err, abuse.ErrCircuitOpen) {
		response.Error(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if errors.Is(err, abuse.ErrDailyLimitReached) {
		response.Error(w, http.StatusTooManyRequests, err.Error())
		return
	}
	if errors.Is(err, abuse.ErrPromptEmpty) ||
		errors.Is(err, abuse.ErrPromptTooShort) ||
		errors.Is(err, abuse.ErrPromptTooLong) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, provider.ErrProviderNotFound) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, provider.ErrProviderFailed) {
		response.Error(w, http.StatusBadGateway, "IaC generation failed: provider error")
		return
	}
	if errors.Is(err, provider.ErrInvalidResponse) {
		response.Error(w, http.StatusUnprocessableEntity, "provider returned invalid response")
		return
	}

	// Anything unrecognized: log at Error so it surfaces via the slog→Sentry
	// bridge. Without this the SPA just sees "internal server error" with no
	// trail — exactly the diagnostic blackhole we hit on cloud-save.
	h.logger.Error("iac generate failed", "error", err)
	response.Error(w, http.StatusInternalServerError, "internal server error")
}

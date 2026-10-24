package diagram

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/oche/diagram/internal/abuse"
	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/internal/provider"
	"github.com/oche/diagram/pkg/response"
	"github.com/oche/diagram/pkg/validate"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Generate handles non-streaming diagram generation.
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

	if err := enforceBYOKey(user, &req); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	result, err := h.service.Generate(r.Context(), user.UserID, string(user.EffectivePlan()), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// GenerateStream handles streaming diagram generation via Server-Sent Events.
func (h *Handler) GenerateStream(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	var req GenerateRequest
	if !validate.DecodeAndValidate(w, r, &req) {
		return
	}

	if err := enforceBYOKey(user, &req); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	// Verify provider exists before switching to SSE
	if _, err := h.service.registry.Get(req.Provider); err != nil {
		h.handleError(w, err)
		return
	}

	// Switch to SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	writeSSE := func(event string, data any) {
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, jsonData)
		flusher.Flush()
	}

	h.service.Stream(
		r.Context(),
		user.UserID,
		string(user.EffectivePlan()),
		req,
		func(chunk provider.StreamChunk) {
			writeSSE("chunk", chunk)
		},
		func(result *GenerateResponse, err error) {
			if err != nil {
				writeSSE("error", map[string]string{"error": err.Error()})
				return
			}
			writeSSE("result", result)
		},
	)
}

// enforceBYOKey blocks non-BYO users from passing their own provider key.
// The frontend already hides the BYO providers, but the API is reachable
// directly — a determined user could curl `api_key=...` against a free
// account and effectively get unlimited generation on our infra dime.
func enforceBYOKey(user *auth.SessionUser, req *GenerateRequest) error {
	if user.EffectivePlan() == auth.PlanBYO {
		return nil
	}
	if req.APIKey == "" {
		return nil
	}
	return errors.New("BYO keys require the BYO plan")
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	// Abuse errors
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

	// Existing errors
	switch {
	case errors.Is(err, ErrEmptyPrompt),
		errors.Is(err, ErrMissingAPIKey),
		errors.Is(err, ErrMissingProvider):
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, provider.ErrProviderNotFound) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, provider.ErrProviderFailed) {
		response.Error(w, http.StatusBadGateway, "diagram generation failed: provider error")
		return
	}

	if errors.Is(err, provider.ErrTruncated) {
		response.Error(w, http.StatusUnprocessableEntity, "diagram exceeded the model output budget — try a more focused prompt")
		return
	}

	if errors.Is(err, provider.ErrInvalidResponse) {
		response.Error(w, http.StatusUnprocessableEntity, "provider returned invalid diagram data")
		return
	}

	var ve *ValidationError
	if errors.As(err, &ve) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]any{
			"error":   "generated diagram failed validation",
			"details": ve.Errors,
		})
		return
	}

	response.Error(w, http.StatusInternalServerError, "internal server error")
}

package cloudsync

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/pkg/response"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service, logger: slog.Default()}
}

// List — GET /api/v1/sync/projects
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	list, err := h.service.List(r.Context(), user.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list projects")
		return
	}
	if list == nil {
		list = []Project{}
	}
	response.JSON(w, http.StatusOK, list)
}

// Get — GET /api/v1/sync/projects/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	p, raw, err := h.service.Load(r.Context(), user.UserID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "project not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to load project")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"project": p,
		"diagram": json.RawMessage(raw),
	})
}

// Put — PUT /api/v1/sync/projects/{id} (also handles create when id == "new")
func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")

	var req SaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			response.Error(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if id != "" && id != "new" {
		req.ID = id
	}

	p, err := h.service.Save(r.Context(), user.UserID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, p)
}

// Preview — GET /api/v1/sync/projects/{id}/preview
// Streams the PNG preview bytes through the backend so the S3 bucket never
// has to be browser-reachable (works with Tailscale / private-VPC Garage).
func (h *Handler) Preview(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	data, err := h.service.LoadPreview(r.Context(), user.UserID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "preview not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to load preview")
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "private, max-age=60")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(data)
}

// CreateShare — POST /api/v1/sync/projects/{id}/share
func (h *Handler) CreateShare(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	token, err := h.service.CreateShare(r.Context(), user.UserID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "project not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "share failed")
		return
	}
	response.JSON(w, http.StatusOK, token)
}

// ListShares — GET /api/v1/sync/projects/{id}/share
func (h *Handler) ListShares(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	shares, err := h.service.ListShares(r.Context(), user.UserID, id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list shares")
		return
	}
	if shares == nil {
		shares = []ShareToken{}
	}
	response.JSON(w, http.StatusOK, shares)
}

// RevokeShare — DELETE /api/v1/sync/projects/{id}/share/{token}
func (h *Handler) RevokeShare(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	token := chi.URLParam(r, "token")
	if err := h.service.RevokeShare(r.Context(), user.UserID, token); err != nil {
		response.Error(w, http.StatusInternalServerError, "revoke failed")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// LoadShared — GET /api/v1/public/share/{token} (no auth)
func (h *Handler) LoadShared(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	p, raw, err := h.service.LoadShared(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "share link invalid or revoked")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to load shared project")
		return
	}
	// Cacheable for 60s — revocation will show a 404 quickly enough.
	w.Header().Set("Cache-Control", "public, max-age=60")
	response.JSON(w, http.StatusOK, map[string]any{
		"project": p,
		"diagram": json.RawMessage(raw),
	})
}

// ListVersions — GET /api/v1/sync/projects/{id}/versions
func (h *Handler) ListVersions(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	versions, err := h.service.ListVersions(r.Context(), user.UserID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "project not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to list versions")
		return
	}
	if versions == nil {
		versions = []VersionSummary{}
	}
	response.JSON(w, http.StatusOK, versions)
}

// RestoreVersion — POST /api/v1/sync/projects/{id}/versions/{versionId}/restore
func (h *Handler) RestoreVersion(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	versionID := chi.URLParam(r, "versionId")
	p, raw, err := h.service.RestoreVersion(r.Context(), user.UserID, id, versionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "version not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "restore failed")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"project": p,
		"diagram": raw,
	})
}

// Delete — DELETE /api/v1/sync/projects/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.service.Delete(r.Context(), user.UserID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "project not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to delete project")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "project not found")
	case errors.Is(err, ErrProjectQuota):
		response.Error(w, http.StatusPaymentRequired, err.Error())
	case errors.Is(err, ErrDiagramTooLarge), errors.Is(err, ErrPreviewTooLarge):
		response.Error(w, http.StatusRequestEntityTooLarge, err.Error())
	case errors.Is(err, ErrInvalidDiagram):
		response.Error(w, http.StatusBadRequest, err.Error())
	default:
		// Unknown / DB-layer failures need to be visible. slog Error records are
		// forwarded to Sentry by the observability wrapper, so this is also how
		// these failures get into the issue tracker.
		h.logger.Error("cloud save failed", "error", err)
		response.Error(w, http.StatusInternalServerError, "save failed")
	}
}

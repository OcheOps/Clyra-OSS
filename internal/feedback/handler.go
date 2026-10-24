package feedback

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type submitRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Category string `json:"category"`
	Message  string `json:"message"`
}

// Submit — POST /api/v1/feedback
//
// Public endpoint. If the caller is logged in we attach their user id and
// override the supplied email with their account email — prevents spoofing
// "<other-user>@<your-domain>" via this form. Anonymous posts get whatever
// email the form gave us, validated only for shape.
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	var req submitRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<15)).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}

	in := Submission{
		Name:      req.Name,
		Email:     req.Email,
		Category:  req.Category,
		Message:   req.Message,
		IP:        clientIP(r),
		UserAgent: r.UserAgent(),
	}
	if u := auth.UserFromContext(r.Context()); u != nil {
		in.UserID = u.UserID
		in.Email = u.Email // trust the session, not the form
		if in.Name == "" {
			in.Name = u.Email
		}
	}

	item, err := h.svc.Submit(r.Context(), in)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not submit feedback")
		return
	}
	response.JSON(w, http.StatusCreated, map[string]string{"id": item.ID, "status": item.Status})
}

// List — GET /api/v1/admin/feedback?status=&limit=&offset=
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	items, total, err := h.svc.List(r.Context(), q.Get("status"), limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "list failed")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

// SetStatus — POST /api/v1/admin/feedback/{id}/status  body: {"status":"open"}
func (h *Handler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "id required")
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.svc.SetStatus(r.Context(), id, req.Status); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "not found")
			return
		}
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "update failed")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

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

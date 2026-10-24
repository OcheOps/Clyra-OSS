package cloudsync

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

var (
	ErrNotFound        = errors.New("project not found")
	ErrProjectQuota    = errors.New("project limit reached for your plan")
	ErrDiagramTooLarge = errors.New("diagram exceeds the per-file size limit")
	ErrPreviewTooLarge = errors.New("preview image exceeds the per-file size limit")
	ErrInvalidDiagram  = errors.New("invalid diagram payload")
)

// PlanResolver returns the quotas the caller should have. Allows the service
// to stay decoupled from the auth package.
type PlanResolver func(userID string) (plan string, ok bool)

// AuditSink mirrors the contract used by other packages so cloudsync can emit
// share-mint / share-revoke events without importing internal/audit.
type AuditSink interface {
	Log(ctx context.Context, event string, fields map[string]any)
}

type Service struct {
	repo    *Repository
	storage Storage
	quotas  func(plan string) Quotas
	resolve PlanResolver
	audit   AuditSink
	logger  *slog.Logger
}

func NewService(repo *Repository, storage Storage, resolve PlanResolver, logger *slog.Logger) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
		quotas:  DefaultQuotas,
		resolve: resolve,
		logger:  logger,
	}
}

// WithAudit attaches the audit sink. Optional.
func (s *Service) WithAudit(a AuditSink) *Service {
	s.audit = a
	return s
}

// Save creates or updates a project, respecting per-plan quotas.
func (s *Service) Save(ctx context.Context, userID string, req SaveRequest) (*Project, error) {
	plan, _ := s.resolve(userID)
	quotas := s.quotas(plan)

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "Untitled Diagram"
	}
	if len(title) > 140 {
		title = title[:140]
	}

	// Re-marshal the diagram to normalize payload + measure size.
	if req.Diagram == nil {
		return nil, ErrInvalidDiagram
	}
	data, err := json.Marshal(req.Diagram)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDiagram, err)
	}
	if len(data) > quotas.MaxDiagramBytes {
		return nil, ErrDiagramTooLarge
	}

	var previewBytes []byte
	if req.PreviewPNGB64 != "" {
		previewBytes, err = decodePNGBase64(req.PreviewPNGB64)
		if err != nil {
			return nil, fmt.Errorf("preview decode: %w", err)
		}
		if len(previewBytes) > quotas.MaxPreviewBytes {
			return nil, ErrPreviewTooLarge
		}
	}

	var rec *Record

	if req.ID != "" {
		existing, err := s.repo.GetByID(ctx, userID, req.ID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			rec = existing
		}
		// If the ID didn't match anything we own, fall through to the create
		// path. Clients may pass a locally-generated draft ID on first save
		// (e.g. "proj-<ts>-<rand>"); treating that as "not found" surfaced as
		// a confusing 404 to the user. Quota still enforced below.
	}
	if rec == nil {
		// New project — enforce project-count quota.
		n, err := s.repo.CountForUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		if n >= quotas.MaxProjects {
			return nil, ErrProjectQuota
		}
		rec = &Record{
			Project: Project{
				ID:      newID(),
				UserID:  userID,
				Title:   title,
				Version: 0, // Insert sets to 1, Update increments.
				Storage: s.storage.Kind(),
				Source:  req.Source,
			},
		}
		if rec.Source == "" {
			rec.Source = SourceApp
		}
	}

	// Write blobs first (so the DB row always references valid keys).
	if s.storage.Kind() == "s3" {
		jsonK, err := s.storage.PutJSON(ctx, userID, rec.ID, data)
		if err != nil {
			return nil, err
		}
		rec.S3Key = jsonK
		rec.Storage = "s3"
		rec.Data = nil

		if len(previewBytes) > 0 {
			previewK, err := s.storage.PutPreview(ctx, userID, rec.ID, previewBytes)
			if err != nil {
				s.logger.Warn("preview upload failed; continuing without preview", "error", err)
			} else {
				rec.PreviewKey = previewK
			}
		}
	} else {
		rec.Data = data
		rec.Storage = "db"
		rec.S3Key = ""
		// DB storage doesn't keep previews — drop.
		rec.PreviewKey = ""
		previewBytes = nil
	}

	rec.Title = title
	rec.ByteSize = len(data)
	rec.PreviewBytes = len(previewBytes)

	// Version is 0 only on records freshly minted in the create branch above
	// (DB rows always have Version >= 1). Use it as the insert vs update signal
	// so a client-provided draft ID that didn't match still flows into Insert.
	if rec.Version == 0 {
		if err := s.repo.Insert(ctx, rec); err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Update(ctx, rec); err != nil {
			return nil, err
		}
	}

	// Re-read so the response carries the DB-assigned version + timestamps.
	fresh, err := s.repo.GetByID(ctx, userID, rec.ID)
	if err == nil && fresh != nil {
		rec = fresh
	}

	// Snapshot this save into the version history (rolling 10). Best-effort:
	// a failure here must not block the user's save.
	if err := s.repo.InsertVersion(ctx, rec.ID, userID, rec.Title, data, 10); err != nil {
		s.logger.Warn("version snapshot failed", "project_id", rec.ID, "error", err)
	}

	out := rec.Project
	out.PreviewURL = s.previewURL(rec)
	out.ByteSize = rec.ByteSize
	out.PreviewBytes = rec.PreviewBytes
	return &out, nil
}

// List returns project metadata (no diagram payload).
func (s *Service) List(ctx context.Context, userID string) ([]Project, error) {
	recs, err := s.repo.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Project, 0, len(recs))
	for i := range recs {
		p := recs[i].Project
		p.ByteSize = recs[i].ByteSize
		p.PreviewBytes = recs[i].PreviewBytes
		p.PreviewURL = s.previewURL(&recs[i])
		out = append(out, p)
	}
	return out, nil
}

// Load returns both metadata and the diagram payload for the editor.
func (s *Service) Load(ctx context.Context, userID, id string) (*Project, json.RawMessage, error) {
	rec, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, nil, err
	}
	if rec == nil {
		return nil, nil, ErrNotFound
	}

	var raw json.RawMessage
	if rec.Storage == "s3" && rec.S3Key != "" {
		body, err := s.storage.GetJSON(ctx, rec.S3Key)
		if err != nil {
			return nil, nil, err
		}
		raw = body
	} else {
		raw = rec.Data
	}

	p := rec.Project
	p.PreviewURL = s.previewURL(rec)
	p.ByteSize = rec.ByteSize
	p.PreviewBytes = rec.PreviewBytes
	return &p, raw, nil
}

// LoadPreview returns the raw PNG bytes for a project's preview image.
// Enforces ownership — used by the HTTP preview proxy.
func (s *Service) LoadPreview(ctx context.Context, userID, id string) ([]byte, error) {
	rec, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if rec == nil || rec.PreviewKey == "" {
		return nil, ErrNotFound
	}
	return s.storage.GetPreview(ctx, rec.PreviewKey)
}

// previewURL picks between a directly-reachable public URL (when
// S3_PUBLIC_BASE_URL is set) and the backend proxy path (the normal case,
// including when S3 is on a private Tailscale/VPC network).
func (s *Service) previewURL(rec *Record) string {
	if rec == nil || rec.PreviewKey == "" {
		return ""
	}
	if url := s.storage.PublicPreviewURL(rec.PreviewKey); url != "" {
		return url
	}
	return "/api/v1/sync/projects/" + rec.ID + "/preview"
}

// ─── Share tokens ────────────────────────────────────────────────────

// CreateShare mints a new public token for a project the user owns.
// Multiple tokens per project are allowed — revoke via RevokeShare.
func (s *Service) CreateShare(ctx context.Context, userID, projectID string) (*ShareToken, error) {
	rec, err := s.repo.GetByID(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	// 24-byte random token → 32-char URL-safe string. Effectively unguessable.
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("share token rand: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	if err := s.repo.CreateShare(ctx, token, projectID, userID); err != nil {
		return nil, err
	}
	if s.audit != nil {
		s.audit.Log(ctx, "share.created", map[string]any{
			"actor_user_id": userID,
			"target_id":     projectID,
		})
	}
	return &ShareToken{Token: token, ProjectID: projectID}, nil
}

func (s *Service) ListShares(ctx context.Context, userID, projectID string) ([]ShareToken, error) {
	return s.repo.ListShares(ctx, userID, projectID)
}

func (s *Service) RevokeShare(ctx context.Context, userID, token string) error {
	if err := s.repo.RevokeShare(ctx, userID, token); err != nil {
		return err
	}
	if s.audit != nil {
		s.audit.Log(ctx, "share.revoked", map[string]any{
			"actor_user_id": userID,
			// We don't log the token itself — just that one was revoked.
		})
	}
	return nil
}

// LoadShared returns a project + diagram by share token. Public — no user
// scoping. Intended for the unauth /api/v1/public/share/{token} endpoint.
func (s *Service) LoadShared(ctx context.Context, token string) (*Project, json.RawMessage, error) {
	rec, err := s.repo.GetShareByToken(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	if rec == nil {
		return nil, nil, ErrNotFound
	}
	var data json.RawMessage
	if s.storage.Kind() == "s3" && rec.S3Key != "" {
		raw, err := s.storage.GetJSON(ctx, rec.S3Key)
		if err != nil {
			return nil, nil, err
		}
		data = raw
	} else {
		data = rec.Data
	}
	out := rec.Project
	out.PreviewURL = "" // No preview URL for public shares (auth-gated).
	return &out, data, nil
}

// ListVersions returns snapshot metadata for a project (newest first).
// Scoped to the owning user.
func (s *Service) ListVersions(ctx context.Context, userID, projectID string) ([]VersionSummary, error) {
	rec, err := s.repo.GetByID(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	return s.repo.ListVersions(ctx, userID, projectID)
}

// RestoreVersion loads a specific snapshot and writes it back as the
// project's current data. A new snapshot of the *current* state is
// captured first so restore is itself reversible.
func (s *Service) RestoreVersion(ctx context.Context, userID, projectID, versionID string) (*Project, json.RawMessage, error) {
	rec, err := s.repo.GetByID(ctx, userID, projectID)
	if err != nil {
		return nil, nil, err
	}
	if rec == nil {
		return nil, nil, ErrNotFound
	}
	data, err := s.repo.GetVersion(ctx, userID, projectID, versionID)
	if err != nil {
		return nil, nil, err
	}
	if data == nil {
		return nil, nil, ErrNotFound
	}

	// Snapshot the current state first so restore can be undone.
	currentData := rec.Data
	if currentData == nil && s.storage.Kind() == "s3" && rec.S3Key != "" {
		currentData, _ = s.storage.GetJSON(ctx, rec.S3Key)
	}
	if currentData != nil {
		_ = s.repo.InsertVersion(ctx, projectID, userID, rec.Title, currentData, 10)
	}

	// Overwrite current data with the snapshot.
	if s.storage.Kind() == "s3" {
		jsonK, err := s.storage.PutJSON(ctx, userID, projectID, data)
		if err != nil {
			return nil, nil, err
		}
		rec.S3Key = jsonK
		rec.Storage = "s3"
		rec.Data = nil
	} else {
		rec.Data = data
	}
	rec.ByteSize = len(data)
	if err := s.repo.Update(ctx, rec); err != nil {
		return nil, nil, err
	}

	fresh, err := s.repo.GetByID(ctx, userID, projectID)
	if err == nil && fresh != nil {
		rec = fresh
	}
	out := rec.Project
	out.PreviewURL = s.previewURL(rec)
	out.ByteSize = rec.ByteSize
	return &out, json.RawMessage(data), nil
}

// Delete removes a project and its blobs.
func (s *Service) Delete(ctx context.Context, userID, id string) error {
	rec, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if rec == nil {
		return ErrNotFound
	}
	if err := s.storage.Delete(ctx, userID, id, rec.S3Key, rec.PreviewKey); err != nil {
		s.logger.Warn("storage delete failed; removing DB row anyway", "error", err)
	}
	return s.repo.Delete(ctx, userID, id)
}

// decodePNGBase64 accepts both raw base64 and "data:image/png;base64,..." forms.
func decodePNGBase64(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "data:") {
		if idx := strings.Index(s, ","); idx >= 0 {
			s = s[idx+1:]
		}
	}
	return base64.StdEncoding.DecodeString(s)
}

func newID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

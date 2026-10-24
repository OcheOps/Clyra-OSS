// Package cloudsync provides opt-in cross-device project storage.
//
// Storage is layered:
//   - Metadata lives in user_projects (Postgres).
//   - Large blobs (diagram JSON, preview image) live in S3 when configured,
//     otherwise inline in the DB.
//
// Quotas are enforced at the service layer: max projects per plan and per-file
// size ceilings (keeps bills predictable in the pre-CAC window).
package cloudsync

import "time"

type ProjectSource string

const (
	SourceApp ProjectSource = "app"
	SourceMCP ProjectSource = "mcp"
)

// Project is the canonical metadata returned to the client. The diagram JSON
// is served by the dedicated Get endpoint (signed URL or inline payload).
type Project struct {
	ID           string        `json:"id"`
	UserID       string        `json:"user_id"`
	Title        string        `json:"title"`
	Version      int           `json:"version"`
	Storage      string        `json:"storage"` // "db" | "s3"
	Source       ProjectSource `json:"source"`
	ByteSize     int           `json:"byte_size"`
	PreviewURL   string        `json:"preview_url,omitempty"`
	PreviewBytes int           `json:"preview_bytes,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// SaveRequest is what the editor POSTs to persist a project.
// `preview_png_b64` is optional — small PNGs base64-encoded.
type SaveRequest struct {
	ID            string        `json:"id,omitempty"` // empty = create, set = update
	Title         string        `json:"title"`
	Diagram       any           `json:"diagram"` // parsed JSON, we re-encode server-side
	PreviewPNGB64 string        `json:"preview_png_b64,omitempty"`
	Source        ProjectSource `json:"-"`
}

// Quotas captures the per-plan storage limits.
type Quotas struct {
	MaxProjects     int
	MaxDiagramBytes int
	MaxPreviewBytes int
}

// DefaultQuotas returns per-plan cloud-sync limits.
//   - Free: tight — 2 projects max. Creates upgrade pressure when users hit
//     the cap; keeps the value of paid tiers tangible.
//   - Pro / BYO: generous — 100 projects, larger diagram + preview sizes.
//
// Adjust in main.go if a deployment wants different thresholds.
func DefaultQuotas(plan string) Quotas {
	switch plan {
	case "pro", "byo":
		return Quotas{MaxProjects: 100, MaxDiagramBytes: 2 << 20, MaxPreviewBytes: 512 << 10}
	default:
		return Quotas{MaxProjects: 2, MaxDiagramBytes: 1 << 20, MaxPreviewBytes: 256 << 10}
	}
}

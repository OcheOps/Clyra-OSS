package cloudsync

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Record is the internal DB-layer shape of a project. The JSON column is
// populated only for storage='db'.
type Record struct {
	Project
	S3Key      string // "" for DB storage
	PreviewKey string // "" when no preview uploaded
	Data       []byte // populated when storage='db'
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CountForUser(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM user_projects WHERE user_id = $1`, userID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("counting projects: %w", err)
	}
	return n, nil
}

// Insert creates a new project row.
func (r *Repository) Insert(ctx context.Context, rec *Record) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_projects
		  (id, user_id, title, version, storage, source, s3_key, preview_key, data, byte_size, preview_bytes, created_at, updated_at)
		 VALUES ($1, $2, $3, 1, $4, $5, NULLIF($6,''), NULLIF($7,''), $8, $9, $10, NOW(), NOW())`,
		rec.ID, rec.UserID, rec.Title, rec.Storage, rec.Source, rec.S3Key, rec.PreviewKey, rec.Data, rec.ByteSize, rec.PreviewBytes,
	)
	if err != nil {
		return fmt.Errorf("inserting project: %w", err)
	}
	return nil
}

// Update writes an existing project row, bumping version + updated_at.
func (r *Repository) Update(ctx context.Context, rec *Record) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE user_projects
		 SET title = $1, storage = $2, s3_key = NULLIF($3,''), preview_key = NULLIF($4,''),
		     data = $5, byte_size = $6, preview_bytes = $7,
		     version = version + 1, updated_at = NOW()
		 WHERE id = $8 AND user_id = $9`,
		rec.Title, rec.Storage, rec.S3Key, rec.PreviewKey, rec.Data, rec.ByteSize, rec.PreviewBytes,
		rec.ID, rec.UserID,
	)
	if err != nil {
		return fmt.Errorf("updating project: %w", err)
	}
	return nil
}

// GetByID fetches a project, ensuring it belongs to userID.
func (r *Repository) GetByID(ctx context.Context, userID, id string) (*Record, error) {
	var rec Record
	var s3Key, previewKey *string
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, title, version, storage, source, s3_key, preview_key, data, byte_size, preview_bytes, created_at, updated_at
		 FROM user_projects WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(
		&rec.ID, &rec.UserID, &rec.Title, &rec.Version, &rec.Storage, &rec.Source,
		&s3Key, &previewKey, &rec.Data, &rec.ByteSize, &rec.PreviewBytes,
		&rec.CreatedAt, &rec.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if s3Key != nil {
		rec.S3Key = *s3Key
	}
	if previewKey != nil {
		rec.PreviewKey = *previewKey
	}
	return &rec, nil
}

func (r *Repository) ListForUser(ctx context.Context, userID string) ([]Record, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, title, version, storage, source, s3_key, preview_key, byte_size, preview_bytes, created_at, updated_at
		 FROM user_projects WHERE user_id = $1 ORDER BY updated_at DESC LIMIT 200`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	defer rows.Close()

	var out []Record
	for rows.Next() {
		var rec Record
		var s3Key, previewKey *string
		if err := rows.Scan(
			&rec.ID, &rec.UserID, &rec.Title, &rec.Version, &rec.Storage, &rec.Source,
			&s3Key, &previewKey, &rec.ByteSize, &rec.PreviewBytes,
			&rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning project: %w", err)
		}
		if s3Key != nil {
			rec.S3Key = *s3Key
		}
		if previewKey != nil {
			rec.PreviewKey = *previewKey
		}
		out = append(out, rec)
	}
	return out, nil
}

func (r *Repository) Delete(ctx context.Context, userID, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_projects WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("deleting project: %w", err)
	}
	return nil
}

// ─── Share tokens ────────────────────────────────────────────────────

type ShareToken struct {
	Token     string    `json:"token"`
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *Repository) CreateShare(ctx context.Context, token, projectID, userID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_project_shares (token, project_id, user_id) VALUES ($1, $2, $3)`,
		token, projectID, userID,
	)
	if err != nil {
		return fmt.Errorf("creating share: %w", err)
	}
	return nil
}

func (r *Repository) ListShares(ctx context.Context, userID, projectID string) ([]ShareToken, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT token, project_id, created_at
		 FROM user_project_shares
		 WHERE project_id = $1 AND user_id = $2
		 ORDER BY created_at DESC`,
		projectID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing shares: %w", err)
	}
	defer rows.Close()
	var out []ShareToken
	for rows.Next() {
		var s ShareToken
		if err := rows.Scan(&s.Token, &s.ProjectID, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *Repository) RevokeShare(ctx context.Context, userID, token string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM user_project_shares WHERE token = $1 AND user_id = $2`,
		token, userID,
	)
	if err != nil {
		return fmt.Errorf("revoking share: %w", err)
	}
	return nil
}

// GetShareByToken resolves a public share token to its project. No
// user scoping — this is the whole point of the public link.
func (r *Repository) GetShareByToken(ctx context.Context, token string) (*Record, error) {
	var rec Record
	var s3Key, previewKey *string
	err := r.pool.QueryRow(ctx,
		`SELECT p.id, p.user_id, p.title, p.version, p.storage, p.source, p.s3_key, p.preview_key, p.data, p.byte_size, p.preview_bytes, p.created_at, p.updated_at
		 FROM user_project_shares s
		 JOIN user_projects p ON p.id = s.project_id
		 WHERE s.token = $1`,
		token,
	).Scan(
		&rec.ID, &rec.UserID, &rec.Title, &rec.Version, &rec.Storage, &rec.Source,
		&s3Key, &previewKey, &rec.Data, &rec.ByteSize, &rec.PreviewBytes,
		&rec.CreatedAt, &rec.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting shared project: %w", err)
	}
	if s3Key != nil {
		rec.S3Key = *s3Key
	}
	if previewKey != nil {
		rec.PreviewKey = *previewKey
	}
	return &rec, nil
}

// ─── Version snapshots ───────────────────────────────────────────────

// VersionSummary is returned in list endpoints — no payload, just metadata.
type VersionSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	ByteSize  int       `json:"byte_size"`
	CreatedAt time.Time `json:"created_at"`
}

// InsertVersion records a snapshot and trims older entries past
// `keepLast`. Trim is best-effort — a rare failure there is far less
// bad than refusing to save.
func (r *Repository) InsertVersion(ctx context.Context, projectID, userID, title string, data []byte, keepLast int) error {
	id := newID()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_project_versions (id, project_id, user_id, title, data, byte_size, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		id, projectID, userID, title, data, len(data),
	)
	if err != nil {
		return fmt.Errorf("inserting version: %w", err)
	}
	if keepLast > 0 {
		// Keep only the `keepLast` most recent; drop anything older.
		_, _ = r.pool.Exec(ctx,
			`DELETE FROM user_project_versions
			 WHERE project_id = $1 AND id NOT IN (
				SELECT id FROM user_project_versions
				WHERE project_id = $1
				ORDER BY created_at DESC LIMIT $2
			 )`,
			projectID, keepLast,
		)
	}
	return nil
}

func (r *Repository) ListVersions(ctx context.Context, userID, projectID string) ([]VersionSummary, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, byte_size, created_at
		 FROM user_project_versions
		 WHERE project_id = $1 AND user_id = $2
		 ORDER BY created_at DESC LIMIT 25`,
		projectID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing versions: %w", err)
	}
	defer rows.Close()
	var out []VersionSummary
	for rows.Next() {
		var v VersionSummary
		if err := rows.Scan(&v.ID, &v.Title, &v.ByteSize, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning version: %w", err)
		}
		out = append(out, v)
	}
	return out, nil
}

// GetVersion returns the snapshot payload for a specific version, scoped
// to the owning user.
func (r *Repository) GetVersion(ctx context.Context, userID, projectID, versionID string) ([]byte, error) {
	var data []byte
	err := r.pool.QueryRow(ctx,
		`SELECT data FROM user_project_versions
		 WHERE id = $1 AND project_id = $2 AND user_id = $3`,
		versionID, projectID, userID,
	).Scan(&data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting version: %w", err)
	}
	return data, nil
}

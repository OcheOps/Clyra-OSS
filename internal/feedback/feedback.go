// Package feedback owns the support inbox. The HTTP surface is:
//
//   POST /api/v1/feedback         (public — anonymous OK)
//   GET  /api/v1/admin/feedback   (admin)
//   POST /api/v1/admin/feedback/{id}/status  (admin)
//
// Submissions land in the `feedback` table and trigger a notification email
// to the operator (best-effort — a failed email never blocks the user-side
// submission).
package feedback

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxNameLen    = 120
	maxEmailLen   = 254
	maxMessageLen = 4000
	minMessageLen = 10
)

var (
	ErrInvalidInput = errors.New("invalid feedback")
	ErrNotFound     = errors.New("feedback not found")
)

// EmailNotifier is the subset of the email service we need. The invite
// service satisfies it via a thin adapter wired in main.go.
type EmailNotifier interface {
	NotifyFeedback(adminEmail string, item Item) error
	Enabled() bool
}

type Service struct {
	pool       *pgxpool.Pool
	mailer     EmailNotifier
	adminEmail string
	logger     *slog.Logger
}

func NewService(pool *pgxpool.Pool, mailer EmailNotifier, adminEmail string, logger *slog.Logger) *Service {
	return &Service{pool: pool, mailer: mailer, adminEmail: adminEmail, logger: logger}
}

// Item is the wire shape returned to admin. user_id is included only for
// admin reads; the public submission endpoint never echoes it back.
type Item struct {
	ID        string    `json:"id"`
	UserID    *string   `json:"user_id,omitempty"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Category  string    `json:"category"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	IP        *string   `json:"ip,omitempty"`
	UserAgent *string   `json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Submission is the validated input from the public form.
type Submission struct {
	UserID    string // empty for anonymous
	Name      string
	Email     string
	Category  string
	Message   string
	IP        string
	UserAgent string
}

// Submit validates, persists, and (best-effort) notifies the operator.
func (s *Service) Submit(ctx context.Context, in Submission) (*Item, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.Message = strings.TrimSpace(in.Message)
	if in.Category == "" {
		in.Category = "general"
	}
	if !validCategory(in.Category) {
		return nil, fmt.Errorf("%w: category", ErrInvalidInput)
	}
	if in.Name == "" || len(in.Name) > maxNameLen {
		return nil, fmt.Errorf("%w: name", ErrInvalidInput)
	}
	if !looksLikeEmail(in.Email) || len(in.Email) > maxEmailLen {
		return nil, fmt.Errorf("%w: email", ErrInvalidInput)
	}
	if len(in.Message) < minMessageLen {
		return nil, fmt.Errorf("%w: message too short", ErrInvalidInput)
	}
	if len(in.Message) > maxMessageLen {
		return nil, fmt.Errorf("%w: message too long", ErrInvalidInput)
	}

	id, err := newID()
	if err != nil {
		return nil, err
	}
	item := &Item{
		ID:        id,
		Name:      in.Name,
		Email:     in.Email,
		Category:  in.Category,
		Message:   in.Message,
		Status:    "new",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if in.UserID != "" {
		uid := in.UserID
		item.UserID = &uid
	}
	if in.IP != "" {
		ip := in.IP
		item.IP = &ip
	}
	if in.UserAgent != "" {
		ua := truncateUA(in.UserAgent)
		item.UserAgent = &ua
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO feedback (id, user_id, name, email, category, message, status, ip, user_agent)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		item.ID, item.UserID, item.Name, item.Email, item.Category, item.Message, item.Status, item.IP, item.UserAgent,
	)
	if err != nil {
		return nil, fmt.Errorf("insert feedback: %w", err)
	}

	// Best-effort notify. A queued / failing email must not stop the user
	// from getting their 200.
	if s.mailer != nil && s.mailer.Enabled() && s.adminEmail != "" {
		if err := s.mailer.NotifyFeedback(s.adminEmail, *item); err != nil {
			s.logger.Warn("feedback notify failed", "id", item.ID, "error", err)
		}
	}
	s.logger.Info("feedback received", "id", item.ID, "category", item.Category, "user_id", in.UserID)
	return item, nil
}

// List returns most-recent feedback items, optionally filtered by status.
func (s *Service) List(ctx context.Context, status string, limit, offset int) ([]Item, int, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	where := ""
	args := []any{}
	if status != "" {
		args = append(args, status)
		where = " WHERE status = $1"
	}

	var total int
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM feedback"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	limitPos := len(args) - 1
	offsetPos := len(args)
	q := fmt.Sprintf(
		`SELECT id, user_id, name, email, category, message, status, ip, user_agent, created_at, updated_at
		 FROM feedback%s
		 ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, limitPos, offsetPos,
	)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]Item, 0, limit)
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.UserID, &it.Name, &it.Email, &it.Category, &it.Message, &it.Status, &it.IP, &it.UserAgent, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

// SetStatus is the admin-side state-machine update.
func (s *Service) SetStatus(ctx context.Context, id, status string) error {
	if status != "new" && status != "open" && status != "closed" {
		return fmt.Errorf("%w: status", ErrInvalidInput)
	}
	tag, err := s.pool.Exec(ctx,
		`UPDATE feedback SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// --- helpers ---

func validCategory(c string) bool {
	switch c {
	case "general", "bug", "feature", "billing", "security", "other":
		return true
	}
	return false
}

func looksLikeEmail(s string) bool {
	at := strings.IndexByte(s, '@')
	if at <= 0 || at >= len(s)-3 {
		return false
	}
	if strings.IndexByte(s[at+1:], '.') < 0 {
		return false
	}
	return true
}

func truncateUA(s string) string {
	if len(s) > 300 {
		return s[:300]
	}
	return s
}

func newID() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "fb_" + hex.EncodeToString(b), nil
}

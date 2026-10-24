package invite

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type RequestStatus string

const (
	StatusPending  RequestStatus = "pending"
	StatusApproved RequestStatus = "approved"
	StatusDeclined RequestStatus = "declined"
)

// Request represents a "I want an invite" submission from the landing page.
type Request struct {
	ID        string        `json:"id"`
	Email     string        `json:"email"`
	Note      string        `json:"note,omitempty"`
	Status    RequestStatus `json:"status"`
	HandledBy *string       `json:"handled_by,omitempty"`
	HandledAt *time.Time    `json:"handled_at,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

// CreateRequest inserts a new pending invite request.
func (r *Repository) CreateRequest(ctx context.Context, email, note string) (*Request, error) {
	id := generateID()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO invite_requests (id, email, note) VALUES ($1, $2, $3)`,
		id, email, note,
	)
	if err != nil {
		return nil, fmt.Errorf("creating invite request: %w", err)
	}
	return &Request{
		ID:        id,
		Email:     email,
		Note:      note,
		Status:    StatusPending,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// HasRecentPendingRequest returns true if the email already has a pending
// request created within the given window. Used to throttle spam submissions.
func (r *Repository) HasRecentPendingRequest(ctx context.Context, email string, within time.Duration) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (
		    SELECT 1 FROM invite_requests
		    WHERE LOWER(email) = LOWER($1)
		      AND status = 'pending'
		      AND created_at > NOW() - $2::interval
		 )`,
		email, fmt.Sprintf("%d seconds", int(within.Seconds())),
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking recent request: %w", err)
	}
	return exists, nil
}

// ListRequests returns invite requests, newest first.
// If status is empty, returns all.
func (r *Repository) ListRequests(ctx context.Context, status RequestStatus) ([]Request, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if status == "" {
		rows, err = r.pool.Query(ctx,
			`SELECT id, email, note, status, handled_by, handled_at, created_at
			 FROM invite_requests ORDER BY created_at DESC LIMIT 200`,
		)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, email, note, status, handled_by, handled_at, created_at
			 FROM invite_requests WHERE status = $1 ORDER BY created_at DESC LIMIT 200`,
			status,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("listing invite requests: %w", err)
	}
	defer rows.Close()

	var result []Request
	for rows.Next() {
		var req Request
		if err := rows.Scan(&req.ID, &req.Email, &req.Note, &req.Status, &req.HandledBy, &req.HandledAt, &req.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning invite request: %w", err)
		}
		result = append(result, req)
	}
	return result, nil
}

// GetRequest fetches a single invite request by ID.
func (r *Repository) GetRequest(ctx context.Context, id string) (*Request, error) {
	var req Request
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, note, status, handled_by, handled_at, created_at
		 FROM invite_requests WHERE id = $1`,
		id,
	).Scan(&req.ID, &req.Email, &req.Note, &req.Status, &req.HandledBy, &req.HandledAt, &req.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting invite request: %w", err)
	}
	return &req, nil
}

// MarkRequestHandled updates the status of a request.
func (r *Repository) MarkRequestHandled(ctx context.Context, id string, status RequestStatus, handledBy string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE invite_requests
		 SET status = $1, handled_by = $2, handled_at = NOW()
		 WHERE id = $3`,
		status, handledBy, id,
	)
	if err != nil {
		return fmt.Errorf("marking request handled: %w", err)
	}
	return nil
}

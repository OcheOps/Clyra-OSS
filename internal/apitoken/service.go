package apitoken

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// MaxTokensPerUser caps the live-token count per user. Protects against
// accidental automation loops and gives the UI a clear "you're at the
// limit, revoke one first" message.
const MaxTokensPerUser = 20

var (
	ErrTokenLimit   = errors.New("token limit reached")
	ErrInvalidName  = errors.New("token name required")
	ErrMalformed    = errors.New("malformed token")
	ErrUnauthorized = errors.New("unauthorized")
)

// Service mints tokens, hashes them, and mediates the repo. Auth handlers
// use it; the middleware talks directly to the repo (single-purpose read).
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Mint generates a new token for the user. Returns the token with Clear
// populated — the ONE moment we hand out the plaintext. Caller (handler)
// serializes it to the response then drops the reference.
func (s *Service) Mint(ctx context.Context, userID, name string) (*Token, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidName
	}
	if len(name) > 64 {
		name = name[:64]
	}

	live, err := s.repo.CountLiveByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if live >= MaxTokensPerUser {
		return nil, ErrTokenLimit
	}

	clear, hash, err := generateToken()
	if err != nil {
		return nil, err
	}

	t := &Token{
		ID:     generateID(),
		UserID: userID,
		Name:   name,
		Source: SourcePAT,
		Clear:  clear,
	}
	if err := s.repo.Create(ctx, t.ID, userID, name, hash, SourcePAT); err != nil {
		return nil, err
	}
	// Re-fetch to populate CreatedAt from the DB default (keeps client-
	// visible timestamp authoritative and consistent with list).
	stored, err := s.repo.GetByHash(ctx, hash)
	if err == nil && stored != nil {
		stored.Clear = clear
		return stored, nil
	}
	return t, nil
}

// List returns the caller's tokens (without cleartext, obviously).
func (s *Service) List(ctx context.Context, userID string) ([]Token, error) {
	return s.repo.ListByUser(ctx, userID)
}

// Revoke marks a token revoked if the caller owns it.
func (s *Service) Revoke(ctx context.Context, userID, tokenID string) error {
	return s.repo.Revoke(ctx, tokenID, userID)
}

// HashForLookup returns the sha256 hex of a cleartext token as the
// middleware needs to compute for its lookup. Public so the middleware
// package doesn't have to re-implement (or import crypto/sha256 just
// for this) — keeps the "token → hash" transform in one place.
func HashForLookup(clear string) (string, error) {
	if !strings.HasPrefix(clear, TokenPrefix) {
		return "", ErrMalformed
	}
	rest := clear[len(TokenPrefix):]
	if len(rest) != TokenBytes*2 {
		return "", ErrMalformed
	}
	// The random portion must be lowercase hex — reject anything else so
	// case-manipulated tokens don't create a lookup ambiguity.
	for _, c := range rest {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return "", ErrMalformed
		}
	}
	sum := sha256.Sum256([]byte(clear))
	return hex.EncodeToString(sum[:]), nil
}

// generateToken returns (cleartext, sha256Hex, error). Cleartext is the
// form the user sees; sha256Hex is what we persist.
func generateToken() (string, string, error) {
	b := make([]byte, TokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("rand: %w", err)
	}
	clear := TokenPrefix + hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(clear))
	return clear, hex.EncodeToString(sum[:]), nil
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

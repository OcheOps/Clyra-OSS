// Package apitoken owns Personal Access Tokens for the MCP server.
// See docs/mcp.md for the auth model.
//
// Cleartext tokens live for exactly one moment: return value of
// Service.Mint. After that we hold only sha256(token). This is what
// keeps a DB read from being a credential theft, and why revoke is
// the only recourse for a lost token — nobody can recover it.
package apitoken

import "time"

// Source distinguishes user-minted PATs from OAuth-issued tokens.
// OAuth ships in phase 4; the column already exists so we don't have
// to touch schema again then.
type Source string

const (
	SourcePAT   Source = "pat"
	SourceOAuth Source = "oauth"
)

// Token is the persisted shape. TokenHash is sha256(clear) hex.
// Clear is populated ONLY on the return of Service.Mint — never
// loaded from the DB, never logged.
type Token struct {
	ID        string     `json:"id"`
	UserID    string     `json:"-"`
	Name      string     `json:"name"`
	Source    Source     `json:"source"`
	LastUsed  *time.Time `json:"last_used,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// Clear is the plaintext token. Set only on Mint's return; JSON-
	// serialized so the API can hand it to the user once.
	Clear string `json:"clear,omitempty"`
}

// TokenPrefix identifies our PATs at a glance in logs / config files.
// Same shape as GitHub/Stripe/Anthropic tokens: prefix_random.
const TokenPrefix = "clyra_pat_"

// TokenBytes is the entropy of the random portion. 32 bytes → 64 hex
// chars → 256 bits. Massive overkill; costs nothing.
const TokenBytes = 32

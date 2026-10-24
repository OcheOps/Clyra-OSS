package abuse

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/oche/diagram/internal/config"
)

var (
	ErrPromptEmpty    = errors.New("invalid_prompt: prompt is required")
	ErrPromptTooLong  = errors.New("invalid_prompt: prompt exceeds maximum length")
	ErrPromptTooShort = errors.New("invalid_prompt: prompt is too short to generate a diagram")
)

// FreePlanPromptMaxLength is the hard cap on prompt length for free users.
// Paid plans (Pro / BYO) get the configured ABUSE_MAX_PROMPT_LENGTH instead,
// which is intended to be larger. Free is fixed in code so the floor cannot be
// raised by env edits — it's a product policy, not a tuning knob.
const FreePlanPromptMaxLength = 2000

// PromptValidator enforces prompt guardrails before any LLM call.
type PromptValidator struct {
	paidMaxLength int
	freeMaxLength int
	minLength     int
}

func NewPromptValidator(cfg config.AbuseConfig) *PromptValidator {
	paidMax := cfg.MaxPromptLength
	if paidMax <= 0 {
		paidMax = FreePlanPromptMaxLength
	}
	if paidMax < FreePlanPromptMaxLength {
		paidMax = FreePlanPromptMaxLength
	}
	return &PromptValidator{
		paidMaxLength: paidMax,
		freeMaxLength: FreePlanPromptMaxLength,
		minLength:     10,
	}
}

// MaxLengthForPlan returns the prompt-length cap for the given plan.
// "free" → fixed FreePlanPromptMaxLength. Anything else → paid cap.
func (v *PromptValidator) MaxLengthForPlan(plan string) int {
	if plan == "free" || plan == "" {
		return v.freeMaxLength
	}
	return v.paidMaxLength
}

// Validate checks the prompt against all guardrails for the given plan.
// Returns a sanitized prompt and an error if validation fails.
func (v *PromptValidator) Validate(prompt, plan string) (string, error) {
	// Sanitize: collapse whitespace, trim
	cleaned := sanitizePrompt(prompt)

	if cleaned == "" {
		return "", ErrPromptEmpty
	}

	charCount := utf8.RuneCountInString(cleaned)

	if charCount < v.minLength {
		return "", ErrPromptTooShort
	}

	maxLen := v.MaxLengthForPlan(plan)
	if charCount > maxLen {
		return "", fmt.Errorf("%w (max %d, got %d)", ErrPromptTooLong, maxLen, charCount)
	}

	return cleaned, nil
}

// sanitizePrompt normalizes whitespace and removes control characters.
func sanitizePrompt(s string) string {
	// Replace tabs, newlines, carriage returns with spaces
	replacer := strings.NewReplacer(
		"\t", " ",
		"\r\n", " ",
		"\r", " ",
		"\n", " ",
	)
	s = replacer.Replace(s)

	// Collapse multiple spaces
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}

	// Strip leading/trailing whitespace
	s = strings.TrimSpace(s)

	// Remove null bytes and other control characters
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= 32 || r == '\n' { // keep printable chars
			b.WriteRune(r)
		}
	}

	return b.String()
}

// PaidPlanMaxTokens is the output-token ceiling for Pro and BYO users. Free
// users stay on the configured cfg.MaxTokens (typically 4096) so a runaway
// generation doesn't burn through Bedrock spend on a non-paying account; paid
// plans get headroom for 50+ node diagrams that would otherwise truncate.
const PaidPlanMaxTokens = 8192

// TokenGuard enforces max_tokens on provider config before the LLM call.
type TokenGuard struct {
	maxTokens int
}

func NewTokenGuard(cfg config.AbuseConfig) *TokenGuard {
	maxTokens := cfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1000
	}
	return &TokenGuard{maxTokens: maxTokens}
}

// ClampTokens ensures the requested token count doesn't exceed the per-plan
// limit. Free → cfg.MaxTokens. Pro/BYO → max(cfg.MaxTokens, PaidPlanMaxTokens)
// so an operator who explicitly raises the env baseline isn't silently capped.
// requestedTokens<=0 falls back to the plan ceiling.
func (g *TokenGuard) ClampTokens(requestedTokens int, plan string) int {
	ceiling := g.maxTokens
	if isPaidPlan(plan) && PaidPlanMaxTokens > ceiling {
		ceiling = PaidPlanMaxTokens
	}
	if requestedTokens <= 0 || requestedTokens > ceiling {
		return ceiling
	}
	return requestedTokens
}

func isPaidPlan(plan string) bool {
	return plan == "pro" || plan == "byo"
}

// MaxTokens returns the configured baseline (free-plan ceiling).
func (g *TokenGuard) MaxTokens() int {
	return g.maxTokens
}

// NodeGuard validates that generated diagrams don't exceed size limits.
type NodeGuard struct {
	maxNodes int
}

func NewNodeGuard(cfg config.AbuseConfig) *NodeGuard {
	maxNodes := cfg.MaxNodes
	if maxNodes <= 0 {
		maxNodes = 25
	}
	return &NodeGuard{maxNodes: maxNodes}
}

// Check returns an error if the diagram exceeds the node limit.
func (g *NodeGuard) Check(nodeCount int) error {
	if nodeCount > g.maxNodes {
		return fmt.Errorf("diagram too large: %d nodes exceeds maximum of %d", nodeCount, g.maxNodes)
	}
	return nil
}

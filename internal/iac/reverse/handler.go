package reverse

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/oche/diagram/internal/abuse"
	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/internal/diagram"
	"github.com/oche/diagram/pkg/response"
)

// MaxBodyBytes caps the HCL payload at 50KB — generous for a single file,
// way under anything that would tax the parser or the daily quota. The
// per-route maxBodySize middleware in internal/server/routes.go applies
// the same cap defensively at the transport layer.
const MaxBodyBytes = 50 * 1024

// IngestRequest is what the SPA sends. Single `source` field on purpose —
// MVP is one TF file at a time. Multi-file support (state.tfstate +
// modules + .tf files) can layer on top later without breaking the contract.
type IngestRequest struct {
	Source string `json:"source"`
}

// IngestResponse mirrors the spec.
//
//   - Diagram is provider.GenerateResponse shape, the same one loadDiagram()
//     already accepts on the frontend. Re-using it means the import flow
//     shares the editor's normalize/render path with AI-generated diagrams.
//   - Warnings is per-resource UX-level notes ("module rendered as group",
//     "3 unsupported types").
//   - Unsupported is the distinct sorted list of HCL types we mapped to
//     placeholder nodes so the UI can surface a "we approximated these"
//     section without re-scanning the diagram.
type IngestResponse struct {
	Diagram     any      `json:"diagram"` // provider.GenerateResponse; typed as any to keep this package leaf
	Warnings    []string `json:"warnings"`
	Unsupported []string `json:"unsupported"`
}

// Handler owns the HTTP endpoint for reverse ingestion. Quota lives on the
// existing abuse.Guard rather than a separate ingest bucket — see comment
// inside Ingest() for why we deviated from the spec on that point.
type Handler struct {
	guard  *abuse.Guard
	logger *slog.Logger
}

func NewHandler(guard *abuse.Guard) *Handler {
	return &Handler{
		guard:  guard,
		logger: slog.Default(),
	}
}

// Ingest handles POST /api/v1/iac/ingest. Returns the rendered diagram on
// success, or a structured error matching the spec's error catalog.
//
// Note on quota: the original spec called for a new `iac_ingest` quota
// kind, but the existing QuotaService is single-bucket per user and adding
// a per-kind dimension means a schema migration for what is, in practice,
// the same abuse-control intent. For MVP we charge an ingest against the
// shared daily bucket (1 ingest = 1 generation). A free user can either
// generate 10 diagrams OR import 10 TF files OR mix and match. We can split
// the buckets later when there's evidence the mix matters.
func (h *Handler) Ingest(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Read up to MaxBodyBytes+1 so we can distinguish "exactly at the cap"
	// from "tried to exceed it." If we read the +1th byte we know the
	// caller blew the budget.
	limited := io.LimitReader(r.Body, MaxBodyBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		h.logger.Warn("iac ingest body read failed", "user_id", user.UserID, "error", err)
		response.Error(w, http.StatusBadRequest, "could not read request body")
		return
	}
	if len(raw) > MaxBodyBytes {
		response.Error(w, http.StatusRequestEntityTooLarge, "too_large")
		return
	}

	var req IngestRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request")
		return
	}
	if req.Source == "" {
		response.Error(w, http.StatusBadRequest, "source required")
		return
	}

	// Quota check. PreCheck wants a "prompt" string for prompt validation;
	// we synthesize one off the source length so the daily counter still
	// increments correctly and prompt validation is satisfied.
	plan := string(user.EffectivePlan())
	if h.guard != nil {
		synthetic := "iac_ingest:" + plan
		if _, err := h.guard.PreCheck(r.Context(), user.UserID, plan, synthetic); err != nil {
			switch {
			case errors.Is(err, abuse.ErrCircuitOpen):
				response.Error(w, http.StatusServiceUnavailable, err.Error())
			case errors.Is(err, abuse.ErrDailyLimitReached):
				response.Error(w, http.StatusTooManyRequests, "daily_limit_reached")
			default:
				h.logger.Warn("iac ingest precheck failed", "user_id", user.UserID, "error", err)
				response.Error(w, http.StatusBadRequest, err.Error())
			}
			return
		}
	}

	// Three-layer pipeline. Each layer is independently testable; this is
	// the only place they're chained together for the real HTTP flow.
	blocks, diags, perr := Parse("import.tf", req.Source)
	if perr != nil {
		// hcl/v2 diagnostics include line numbers; surface the first one so
		// the user can fix it without us logging a stacktrace.
		msg := "invalid_hcl"
		if diags != nil && len(diags.Errs()) > 0 {
			msg = diags.Errs()[0].Error()
		}
		response.Error(w, http.StatusBadRequest, msg)
		return
	}
	if len(blocks) == 0 {
		response.Error(w, http.StatusBadRequest, "no architecture resources found")
		return
	}

	result := Build(blocks)
	diag := ToDiagram(result)

	// Run the same post-processing the AI-generated path uses, so an
	// imported diagram is indistinguishable from a generated one once it
	// hits the editor: Normalize canonicalizes types, Salvage drops any
	// dangling refs, Validate enforces the schema.
	diagram.Normalize(diag)
	salvage := diagram.Salvage(diag)
	if salvage.DroppedEdges > 0 || salvage.OrphanGroupRef > 0 {
		h.logger.Info("ingest salvage",
			"user_id", user.UserID,
			"dropped_edges", salvage.DroppedEdges,
			"orphan_group_refs", salvage.OrphanGroupRef,
		)
	}
	if ve := diagram.Validate(diag); ve != nil {
		// Validate fires on hard schema breaks (missing IDs, duplicates).
		// If we get here something is wrong with the converter — log loudly,
		// fail the request. Don't surface raw validator errors to the user.
		h.logger.Error("ingest produced invalid diagram",
			"user_id", user.UserID,
			"errors", ve.Errors,
		)
		response.Error(w, http.StatusInternalServerError, "internal")
		return
	}

	// PostSuccess increments the daily counter + records latency = 0 since
	// no provider was involved. node count = total nodes in the response.
	if h.guard != nil {
		_ = h.guard.PostSuccess(r.Context(), user.UserID, "iac_ingest", len(diag.Nodes), 0)
	}

	h.logger.Info("iac ingest succeeded",
		"user_id", user.UserID,
		"input_bytes", len(req.Source),
		"resources", len(result.Graph.Resources),
		"edges", len(diag.Edges),
		"modules", len(result.Modules),
		"unsupported", len(result.UnsupportedKinds),
	)

	response.JSON(w, http.StatusOK, IngestResponse{
		Diagram:     diag,
		Warnings:    result.Warnings,
		Unsupported: result.UnsupportedKinds,
	})
}

// Package reverse turns Terraform HCL into the same ResourceGraph our
// forward (graph → IaC) pipeline already understands. Inverse direction
// of internal/iac/render/terraform.go.
//
// This file is the parser layer ONLY — it walks blocks, captures raw
// attribute expressions as strings, and emits a RawBlock list. It does
// NOT evaluate variables, look up modules, or expand for_each / count.
// The day-2 builder consumes RawBlocks and produces a ResourceGraph;
// keeping the layers split makes both independently testable and lets
// us swap the parser if hcl/v2 ever proves wrong for the job.
package reverse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// BlockKind enumerates the top-level HCL block types we look at. Anything
// outside this set is ignored (terraform, provider, locals, variable,
// output, etc. — none of those describe architecture nodes).
type BlockKind string

const (
	BlockResource BlockKind = "resource"
	BlockData     BlockKind = "data"
	BlockModule   BlockKind = "module"
)

// RawBlock is what the parser emits. It captures the block identity plus
// every top-level attribute as a raw source-text expression. The builder
// later interprets these — including scanning the expression text for
// cross-resource references like `aws_subnet.public.id`.
type RawBlock struct {
	Kind BlockKind
	// Type is the HCL block's first label — `aws_vpc`, `aws_subnet`, etc.
	// For BlockModule there is no type; the field is empty.
	Type string
	// Name is the user-chosen identifier — `main`, `public`, etc. For
	// BlockResource and BlockData this is the second label. For BlockModule
	// it is the first (and only) label.
	Name string
	// Attrs maps each top-level attribute name to its raw expression
	// source. Nested blocks (e.g. a `lifecycle {}` inside a resource) are
	// captured under their block-type key with the *full source text* of
	// the nested block as the value — good enough to detect references
	// from within them, which is the only signal the builder needs.
	Attrs map[string]string
	// Range is the source location of the block header, used in error
	// messages so callers can point users at the right line.
	Range hcl.Range
}

// Parse takes Terraform source text and returns the RawBlocks plus any
// diagnostic messages. A returned non-nil error means the source was
// unparseable; diagnostics on a nil error are warnings (e.g. unknown
// block types). Callers should still try to consume the partial result
// — a TF file with one broken resource shouldn't kill the whole import.
func Parse(filename, source string) ([]RawBlock, hcl.Diagnostics, error) {
	file, diags := hclsyntax.ParseConfig([]byte(source), filename, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, diags, fmt.Errorf("hcl parse: %s", diags.Error())
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, diags, fmt.Errorf("unexpected body type %T", file.Body)
	}

	var out []RawBlock
	for _, b := range body.Blocks {
		switch b.Type {
		case "resource":
			if len(b.Labels) < 2 {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Summary:  "resource block missing labels",
					Detail:   "expected `resource \"type\" \"name\"`",
					Subject:  b.TypeRange.Ptr(),
				})
				continue
			}
			out = append(out, blockToRaw(b, BlockResource, b.Labels[0], b.Labels[1], source))
		case "data":
			if len(b.Labels) < 2 {
				continue
			}
			out = append(out, blockToRaw(b, BlockData, b.Labels[0], b.Labels[1], source))
		case "module":
			if len(b.Labels) < 1 {
				continue
			}
			out = append(out, blockToRaw(b, BlockModule, "", b.Labels[0], source))
		default:
			// terraform {}, provider {}, locals {}, variable {}, output {}:
			// not architecture. Silently skip.
			continue
		}
	}
	return out, diags, nil
}

// blockToRaw extracts attributes + nested-block source slices from a single
// hclsyntax.Block. Top-level attributes become Attrs[name] = <expr text>.
// Nested blocks become Attrs[blockType] = <full block source slice>, so a
// `vpc_config { subnet_ids = [...] }` is searchable for cross-references
// without our needing per-resource-kind knowledge at parse time.
func blockToRaw(b *hclsyntax.Block, kind BlockKind, typ, name, source string) RawBlock {
	attrs := make(map[string]string, len(b.Body.Attributes)+len(b.Body.Blocks))

	for n, attr := range b.Body.Attributes {
		attrs[n] = sliceSource(source, attr.Expr.Range())
	}
	for _, nested := range b.Body.Blocks {
		// If the same nested block name appears multiple times (e.g. two
		// `ingress` blocks on a security group), concatenate their source so
		// reference-scanning catches refs inside all of them. The builder
		// can re-parse the slice if it needs per-block structure later.
		key := nested.Type
		if existing, ok := attrs[key]; ok {
			attrs[key] = existing + "\n" + sliceSource(source, nested.Range())
		} else {
			attrs[key] = sliceSource(source, nested.Range())
		}
	}

	return RawBlock{
		Kind:  kind,
		Type:  typ,
		Name:  name,
		Attrs: attrs,
		Range: b.TypeRange,
	}
}

// sliceSource returns the substring of source covered by r. HCL ranges are
// byte offsets, so we can index directly. Clamped to the source length to
// stay defensive against parser oddities.
func sliceSource(source string, r hcl.Range) string {
	start := r.Start.Byte
	end := r.End.Byte
	if start < 0 || start >= len(source) {
		return ""
	}
	if end > len(source) {
		end = len(source)
	}
	if end <= start {
		return ""
	}
	return source[start:end]
}

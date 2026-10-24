package reverse

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/oche/diagram/internal/iac/graph"
)

// BuildResult is what the builder produces. It's a superset of
// graph.ResourceGraph because reverse-ingestion deals with two concepts the
// forward graph doesn't model: modules (rendered as groups, not nodes) and
// unsupported HCL types (rendered as placeholder nodes per the locked spec).
//
// The day-3 diagram converter consumes this and produces
// provider.GenerateResponse — at that point modules become Groups and
// unsupported resources become "aws.unknown" nodes with the original type
// preserved in metadata.
type BuildResult struct {
	Graph            *graph.ResourceGraph
	Modules          []ModuleRef
	Warnings         []string
	UnsupportedKinds []string // sorted, distinct HCL types we couldn't map
}

// ModuleRef captures a `module "x" { ... }` block. We don't expand modules —
// MVP scope is "render as a labeled box." Source + attrs are kept for the
// diagram converter (label, tooltip) and future tooling.
type ModuleRef struct {
	Name   string
	Source string
	Attrs  map[string]string
}

// refRegex matches the canonical Terraform reference syntax
// `aws_<type>.<name>` and `module.<name>`. The third capture group is the
// optional attribute suffix (`.id`, `.arn`, …) — we discard it because edges
// are between resources, not attribute values. Matches are case-sensitive
// (Terraform identifiers are).
//
// Quoted strings inside HCL still keep refs as bare identifiers, so we don't
// need to strip quoting — `vpc_id = aws_vpc.main.id` and
// `vpc_id = "${aws_vpc.main.id}"` both produce the same match.
var refRegex = regexp.MustCompile(`\b((?:aws_[a-z0-9_]+|module))\.([a-z][a-z0-9_-]*)(?:\.([a-z][a-z0-9_]*))?`)

// Build turns parser output into a BuildResult.
//
// Algorithm:
//  1. First pass — for every block, mint a resource ID. Known HCL types get
//     their canonical ResourceKind; unknown types are still added with the
//     raw type cast to ResourceKind so edges can reference them. Modules
//     get a synthetic ID prefixed `module.` and live in the Modules list.
//  2. Second pass — for every attribute (and nested-block source slice),
//     extract refs via refRegex and emit one Edge per unique (from, to).
//     Self-refs and refs to names not in the resource map are silently
//     skipped (could be a typo, a variable, a data source — none of which
//     are architecture edges).
func Build(blocks []RawBlock) *BuildResult {
	known := graph.Known()
	resources := make([]graph.Resource, 0, len(blocks))
	modules := make([]ModuleRef, 0)
	unsupportedSet := map[string]struct{}{}

	// idMap keys: "<hcl_type>.<name>" → resource ID. We use the same key
	// shape the refRegex produces so edge inference is one direct lookup.
	// Modules use the key shape "module.<name>".
	idMap := map[string]string{}

	for _, b := range blocks {
		switch b.Kind {
		case BlockResource:
			id := newID(b.Type, b.Name)
			key := b.Type + "." + b.Name
			idMap[key] = id

			kind := graph.ResourceKind(b.Type)
			if !known[kind] {
				// Unsupported type — keep it as a resource so users see what
				// they have and refs to it still draw, per locked decision (c).
				// The diagram converter maps unknown kinds to placeholder nodes.
				unsupportedSet[b.Type] = struct{}{}
			}

			res := graph.Resource{
				ID:   id,
				Kind: kind,
				// Drop the raw attr expressions onto Props so the converter
				// has them for tooltips / "view source" later. Casting to any
				// because Props's value type is any, not string.
				Props: rawAttrsToProps(b.Attrs),
			}
			resources = append(resources, res)

		case BlockModule:
			id := "module." + b.Name
			idMap[id] = id
			modules = append(modules, ModuleRef{
				Name:   b.Name,
				Source: stripQuotes(b.Attrs["source"]),
				Attrs:  b.Attrs,
			})

		case BlockData:
			// Data sources aren't infrastructure — they're lookups. Skip.
			// We do NOT register them in idMap, which means refs *to* a
			// data block (e.g. `data.aws_ami.x.id`) won't produce an edge.
			// That's the correct behavior: the architectural dependency is
			// on whatever resource consumes the data, not the lookup itself.
			continue
		}
	}

	// Second pass: edges. Use a set keyed on "<from>->|<to>|<relation>" to
	// dedupe — `subnet_ids = [aws_subnet.a.id, aws_subnet.a.arn]` produces
	// two regex matches but should only emit one edge.
	type edgeKey struct{ from, to string }
	edgeSet := map[edgeKey]graph.Edge{}

	for _, b := range blocks {
		if b.Kind != BlockResource {
			continue
		}
		// Source the from-ID by key rather than the loop index — the
		// resources slice is shorter than blocks (data blocks skipped,
		// modules live in their own list) so positional indexing drifts.
		fromID, ok := idMap[b.Type+"."+b.Name]
		if !ok {
			continue
		}

		for _, exprText := range b.Attrs {
			for _, m := range refRegex.FindAllStringSubmatch(exprText, -1) {
				targetKey := m[1] + "." + m[2]
				toID, ok := idMap[targetKey]
				if !ok || toID == fromID {
					continue
				}
				key := edgeKey{from: fromID, to: toID}
				if _, exists := edgeSet[key]; exists {
					continue
				}
				edgeSet[key] = graph.Edge{
					From:     fromID,
					To:       toID,
					Relation: graph.RelationReferences,
				}
			}
		}
	}

	// Stable order: sort edges by (from, to) so test fixtures are
	// deterministic. Resources are already in input order.
	edges := make([]graph.Edge, 0, len(edgeSet))
	for _, e := range edgeSet {
		edges = append(edges, e)
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})

	unsupported := make([]string, 0, len(unsupportedSet))
	for k := range unsupportedSet {
		unsupported = append(unsupported, k)
	}
	sort.Strings(unsupported)

	warnings := []string{}
	if len(modules) > 0 {
		warnings = append(warnings, fmt.Sprintf(
			"%d module block(s) rendered as opaque groups — contents not expanded", len(modules)))
	}
	if len(unsupported) > 0 {
		warnings = append(warnings, fmt.Sprintf(
			"%d unsupported HCL type(s) rendered as placeholder nodes: %s",
			len(unsupported), strings.Join(unsupported, ", ")))
	}

	return &BuildResult{
		Graph: &graph.ResourceGraph{
			Version:   graph.SchemaVersion,
			Cloud:     graph.CloudAWS,
			Resources: resources,
			Edges:     edges,
			Metadata: graph.Metadata{
				StackName:   "imported",
				Description: "Reverse-ingested from Terraform",
			},
		},
		Modules:          modules,
		Warnings:         warnings,
		UnsupportedKinds: unsupported,
	}
}

// newID mints a deterministic, human-readable resource ID. Reusing the HCL
// label (`aws_vpc.main` → "vpc-main") keeps the diagram readable when users
// inspect node IDs. We don't need uniqueness collision handling because TF
// already forbids duplicate `<type>.<name>` pairs.
func newID(hclType, name string) string {
	short := strings.TrimPrefix(hclType, "aws_")
	return short + "-" + name
}

// rawAttrsToProps converts the parser's map[string]string into the Props
// map[string]any the graph schema expects. We keep values as strings — the
// converter / UI surface them as raw expression text in tooltips.
func rawAttrsToProps(attrs map[string]string) map[string]any {
	if len(attrs) == 0 {
		return nil
	}
	out := make(map[string]any, len(attrs))
	for k, v := range attrs {
		out[k] = v
	}
	return out
}

// stripQuotes removes a leading/trailing double-quote pair if present.
// HCL string literals are captured by the parser including quotes, so the
// module `source = "./x"` attribute comes through as `"./x"`. We want the
// raw path for display.
func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

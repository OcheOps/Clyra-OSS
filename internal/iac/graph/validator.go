package graph

import (
	"errors"
	"fmt"
	"regexp"
)

var idRe = regexp.MustCompile(`^[a-z][a-z0-9_]{1,62}$`)

type ValidationError struct {
	Errors []string
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("graph validation failed: %d error(s): %v", len(v.Errors), v.Errors)
}

func Validate(g *ResourceGraph) error {
	ve := &ValidationError{}

	if g.Version != SchemaVersion {
		ve.Errors = append(ve.Errors, fmt.Sprintf("version must be %q, got %q", SchemaVersion, g.Version))
	}
	if g.Cloud != CloudAWS {
		ve.Errors = append(ve.Errors, fmt.Sprintf("cloud must be %q for v1, got %q", CloudAWS, g.Cloud))
	}
	if g.Region == "" {
		ve.Errors = append(ve.Errors, "region is required")
	}
	if g.Metadata.StackName == "" {
		ve.Errors = append(ve.Errors, "metadata.stack_name is required")
	}

	seen := map[string]bool{}
	known := Known()

	for i, r := range g.Resources {
		if !idRe.MatchString(r.ID) {
			ve.Errors = append(ve.Errors, fmt.Sprintf("resources[%d].id %q: must match %s", i, r.ID, idRe))
		}
		if seen[r.ID] {
			ve.Errors = append(ve.Errors, fmt.Sprintf("resources[%d].id %q: duplicate", i, r.ID))
		}
		seen[r.ID] = true

		if !known[r.Kind] {
			ve.Errors = append(ve.Errors, fmt.Sprintf("resources[%d].kind %q: unsupported", i, r.Kind))
			continue
		}
		if err := ValidateProps(r.Kind, r.Props); err != nil {
			ve.Errors = append(ve.Errors, fmt.Sprintf("resources[%d] (%s): %v", i, r.ID, err))
		}
	}

	for i, e := range g.Edges {
		if !seen[e.From] {
			ve.Errors = append(ve.Errors, fmt.Sprintf("edges[%d].from %q: unknown resource", i, e.From))
		}
		if !seen[e.To] {
			ve.Errors = append(ve.Errors, fmt.Sprintf("edges[%d].to %q: unknown resource", i, e.To))
		}
	}

	for i, r := range g.Resources {
		for _, d := range r.DependsOn {
			if !seen[d] {
				ve.Errors = append(ve.Errors, fmt.Sprintf("resources[%d].depends_on %q: unknown resource", i, d))
			}
		}
	}

	for i, o := range g.Outputs {
		if !seen[o.From] {
			ve.Errors = append(ve.Errors, fmt.Sprintf("outputs[%d].from %q: unknown resource", i, o.From))
		}
	}

	if len(ve.Errors) > 0 {
		return ve
	}
	return nil
}

var ValidateProps = func(kind ResourceKind, props map[string]any) error {
	return errors.New("graph/props package must be imported for prop validation")
}

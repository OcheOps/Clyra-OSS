package render

import (
	"fmt"
	"sort"

	"github.com/oche/diagram/internal/iac/graph"
)

func TopoSort(g *graph.ResourceGraph) ([]graph.Resource, error) {
	byID := make(map[string]graph.Resource, len(g.Resources))
	edges := make(map[string]map[string]bool, len(g.Resources))
	indeg := make(map[string]int, len(g.Resources))

	for _, r := range g.Resources {
		byID[r.ID] = r
		edges[r.ID] = map[string]bool{}
		indeg[r.ID] = 0
	}

	addEdge := func(from, to string) {
		if from == "" || to == "" || edges[from] == nil {
			return
		}
		if !edges[from][to] {
			edges[from][to] = true
			indeg[to]++
		}
	}

	for _, r := range g.Resources {
		for _, d := range r.DependsOn {
			if _, ok := byID[d]; ok {
				addEdge(d, r.ID)
			}
		}
	}
	for _, e := range g.Edges {
		if e.Relation == graph.RelationAttachesTo || e.Relation == graph.RelationRoutesTo {
			addEdge(e.To, e.From)
		}
	}

	var out []graph.Resource
	ready := make([]string, 0, len(indeg))
	for id, d := range indeg {
		if d == 0 {
			ready = append(ready, id)
		}
	}
	for len(ready) > 0 {
		sort.Strings(ready)
		id := ready[0]
		ready = ready[1:]
		out = append(out, byID[id])
		for next := range edges[id] {
			indeg[next]--
			if indeg[next] == 0 {
				ready = append(ready, next)
			}
		}
	}
	if len(out) != len(g.Resources) {
		return nil, fmt.Errorf("cycle detected in resource graph")
	}
	return out, nil
}

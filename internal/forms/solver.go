package forms

import (
	"fmt"
	"slices"
	"strings"
)

// DependencyGraph builds and resolves the DAG of all form fields.
type DependencyGraph struct {
	registry *Registry
	// adjacency list: key -> list of keys it depends on
	edges map[string][]string
	// all known field keys
	nodes map[string]bool
}

// NewDependencyGraph creates a new DependencyGraph backed by the given registry.
func NewDependencyGraph(registry *Registry) *DependencyGraph {
	return &DependencyGraph{
		registry: registry,
		edges:    make(map[string][]string),
		nodes:    make(map[string]bool),
	}
}

// Build constructs the dependency graph from all registered forms.
// Returns an error if a cycle is detected.
func (g *DependencyGraph) Build() error {
	g.edges = make(map[string][]string)
	g.nodes = make(map[string]bool)

	for _, form := range g.registry.AllForms() {
		for _, field := range form.Fields {
			key := FieldKey(form.ID, field.Line)
			g.nodes[key] = true

			// Resolve dependencies. Wildcard deps expand to all matching keys.
			var resolved []string
			for _, dep := range field.DependsOn {
				if strings.Contains(dep, "*") {
					// Expand wildcard to all matching registered field keys
					for node := range g.nodes {
						if matchWildcard(dep, node) {
							resolved = append(resolved, node)
						}
					}
					// Also check nodes added later — we do a second pass below
				} else {
					resolved = append(resolved, dep)
				}
			}
			g.edges[key] = resolved
		}
	}

	// Second pass: expand wildcards again now that all nodes are known.
	for _, form := range g.registry.AllForms() {
		for _, field := range form.Fields {
			key := FieldKey(form.ID, field.Line)
			for _, dep := range field.DependsOn {
				if strings.Contains(dep, "*") {
					existing := make(map[string]bool)
					for _, e := range g.edges[key] {
						existing[e] = true
					}
					for node := range g.nodes {
						if matchWildcard(dep, node) && !existing[node] {
							g.edges[key] = append(g.edges[key], node)
						}
					}
				}
			}
		}
	}

	// Check for cycles
	_, err := g.TopologicalSort()
	return err
}

// TopologicalSort returns all field keys in dependency order using Kahn's algorithm.
// Fields with no dependencies come first. Returns an error if a cycle is detected.
func (g *DependencyGraph) TopologicalSort() ([]string, error) {
	// Compute in-degree for each node
	inDegree := make(map[string]int)
	for node := range g.nodes {
		inDegree[node] = 0
	}

	// Build reverse adjacency: for each edge (key depends on dep),
	// dep -> key means dep must come before key
	reverseAdj := make(map[string][]string)
	for key, deps := range g.edges {
		for _, dep := range deps {
			// Only count edges to known nodes
			if g.nodes[dep] {
				inDegree[key]++
				reverseAdj[dep] = append(reverseAdj[dep], key)
			}
		}
	}

	// Start with nodes that have zero in-degree
	var queue []string
	for node := range g.nodes {
		if inDegree[node] == 0 {
			queue = append(queue, node)
		}
	}

	// Sort queue for deterministic output
	slices.Sort(queue)

	var sorted []string
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		sorted = append(sorted, node)

		for _, dependent := range reverseAdj[node] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
				slices.Sort(queue)
			}
		}
	}

	if len(sorted) != len(g.nodes) {
		// Find nodes involved in cycle for error message
		var cycleNodes []string
		for node, deg := range inDegree {
			if deg > 0 {
				cycleNodes = append(cycleNodes, node)
			}
		}
		slices.Sort(cycleNodes)
		return nil, fmt.Errorf("cycle detected involving fields: %s", strings.Join(cycleNodes, ", "))
	}

	return sorted, nil
}

// MissingInputs returns all UserInput field keys that don't have values in the
// provided map.
func (g *DependencyGraph) MissingInputs(provided map[string]float64) []string {
	var missing []string
	for _, form := range g.registry.AllForms() {
		for _, field := range form.Fields {
			if field.Type == UserInput {
				key := FieldKey(form.ID, field.Line)
				if _, ok := provided[key]; !ok {
					missing = append(missing, key)
				}
			}
		}
	}
	slices.Sort(missing)
	return missing
}

// Solve resolves all Computed fields given the provided UserInput values.
// Returns all field values (inputs + computed). Returns an error if required
// UserInput values are missing.
func (g *DependencyGraph) Solve(inputs map[string]float64, strInputs map[string]string, taxYear int) (map[string]float64, error) {
	// Check for missing required inputs
	missing := g.MissingInputs(inputs)
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required inputs: %s", strings.Join(missing, ", "))
	}

	order, err := g.TopologicalSort()
	if err != nil {
		return nil, err
	}

	// Initialize result with provided inputs
	result := make(map[string]float64)
	for k, v := range inputs {
		result[k] = v
	}

	strResult := make(map[string]string)
	for k, v := range strInputs {
		strResult[k] = v
	}

	// Process fields in topological order
	for _, key := range order {
		if _, hasValue := result[key]; hasValue {
			continue // already provided as input
		}

		_, field, err := g.registry.GetField(key)
		if err != nil {
			continue
		}

		switch field.Type {
		case UserInput:
			// Should already be in result; if not it's an error caught above
			continue
		case Computed, Lookup, FederalRef, PriorYear:
			dv := NewDepValues(result, strResult, taxYear)
			if field.Compute != nil {
				result[key] = field.Compute(dv)
			}
			if field.ComputeStr != nil {
				strResult[key] = field.ComputeStr(dv)
			}
		}
	}

	return result, nil
}


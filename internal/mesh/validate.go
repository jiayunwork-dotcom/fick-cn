package mesh

import "fmt"

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("mesh: %s: %s", e.Field, e.Message)
}

func (g Grid) Validate() error {
	if g.Nodes < 3 {
		return &ValidationError{
			Field:   "nodes",
			Message: fmt.Sprintf("grid needs at least 3 nodes, got %d", g.Nodes),
		}
	}
	if !(g.Length > 0) {
		return &ValidationError{
			Field:   "length",
			Message: fmt.Sprintf("rod length must be positive, got %g", g.Length),
		}
	}
	if !(g.Dx > 0) {
		return &ValidationError{
			Field:   "spacing",
			Message: fmt.Sprintf("node spacing must be positive, got %g", g.Dx),
		}
	}
	expected := g.Length / float64(g.Nodes-1)
	rel := absDiff(g.Dx, expected) / expected
	if rel > 1e-9 {
		return &ValidationError{
			Field: "spacing",
			Message: fmt.Sprintf("spacing %g is inconsistent with length %g and nodes %d (expected %g)",
				g.Dx, g.Length, g.Nodes, expected),
		}
	}
	return nil
}

func (g Grid) Require(prefix string) error {
	if err := g.Validate(); err != nil {
		return fmt.Errorf("%s: %w", prefix, err)
	}
	return nil
}

func absDiff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}

func (g Grid) CheckUniform() (bool, float64) {
	if g.Nodes < 2 {
		return true, 0
	}
	maxRel := 0.0
	for i := 1; i < g.Nodes; i++ {
		dx := (g.Position(i) - g.Position(i-1))
		rel := absDiff(dx, g.Dx) / g.Dx
		if rel > maxRel {
			maxRel = rel
		}
	}
	return maxRel < 1e-12, maxRel
}

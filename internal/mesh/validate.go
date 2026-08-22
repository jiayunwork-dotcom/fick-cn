package mesh

import "fmt"

// ValidationError carries the offending quantity and a human-readable
// description so a CLI can print a precise failure instead of a generic one.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("mesh: %s: %s", e.Field, e.Message)
}

// Validate runs every structural check on the grid and returns the first
// problem found, or nil when the grid is usable.
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
	// Guard against an inconsistent grid that bypassed New (zero-value Grid
	// used directly as a struct literal).
	expected := g.Length / float64(g.Nodes-1)
	rel := absDiff(g.Dx, expected) / expected
	if rel > 1e-9 {
		return &ValidationError{
			Field:   "spacing",
			Message: fmt.Sprintf("spacing %g is inconsistent with length %g and nodes %d (expected %g)",
				g.Dx, g.Length, g.Nodes, expected),
		}
	}
	return nil
}

// Require checks g.Validate and wraps the error with a caller-supplied prefix.
func (g Grid) Require(prefix string) error {
	if err := g.Validate(); err != nil {
		return fmt.Errorf("%s: %w", prefix, err)
	}
	return nil
}

// absDiff returns the absolute difference between two floats.
func absDiff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}

// CheckUniform returns whether every adjacent pair of nodes is separated by
// exactly Dx within a relative tolerance.
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

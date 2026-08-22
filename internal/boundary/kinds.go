// Package boundary models the two ends of the diffusion rod.  Each end is
// either Dirichlet (the concentration is pinned to a fixed value) or
// Neumann (the derivative vanishes, so no mass crosses the end).
//
// The package knows nothing about time stepping; it answers three questions
// a solver needs: how the boundary is declared (Parse), how it enters the
// tridiagonal rows (Row), and how much mass crosses it per step (Flux).
package boundary

import (
	"fmt"
	"strings"
)

// Kind is the type of condition applied at one end of the rod.
type Kind int

const (
	// Dirichlet pins the concentration at the boundary node to a fixed value.
	Dirichlet Kind = iota
	// Neumann enforces a zero derivative, i.e. a no-flux end.
	Neumann
)

// Boundary is the condition on one end of the rod.
type Boundary struct {
	Kind Kind
	// Value is the pinned concentration for a Dirichlet end; unused for
	// Neumann.
	Value float64
}

// ParseKind maps a case-insensitive string ("dirichlet", "neumann") onto a
// Kind.  Anything else is an error so typos in a JSON problem cannot silently
// change the physics.
func ParseKind(s string) (Kind, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "dirichlet", "d", "fixed":
		return Dirichlet, nil
	case "neumann", "n", "no-flux", "noflux", "zero-flux", "insulated":
		return Neumann, nil
	default:
		return 0, fmt.Errorf("boundary: unknown condition %q (want dirichlet or neumann)", s)
	}
}

// New builds a boundary from a kind name and an optional value.  The value is
// validated for Dirichlet ends (must be finite) and ignored for Neumann ends.
func New(kindName string, value float64) (Boundary, error) {
	k, err := ParseKind(kindName)
	if err != nil {
		return Boundary{}, err
	}
	b := Boundary{Kind: k, Value: value}
	if err := b.Validate(); err != nil {
		return Boundary{}, err
	}
	return b, nil
}

// Validate checks the boundary's internal consistency.  A Dirichlet value
// must be finite; a NaN or infinite pinned concentration would poison the
// whole tridiagonal solve.
func (b Boundary) Validate() error {
	if b.Kind == Dirichlet {
		if b.Value != b.Value { // NaN
			return fmt.Errorf("boundary: Dirichlet value is NaN")
		}
		if b.Value > 1e300 || b.Value < -1e300 {
			return fmt.Errorf("boundary: Dirichlet value %g is not finite", b.Value)
		}
	}
	return nil
}

// IsDirichlet reports whether the boundary pins the concentration.
func (b Boundary) IsDirichlet() bool { return b.Kind == Dirichlet }

// IsNoFlux reports whether the boundary is an insulated, no-flux end.
func (b Boundary) IsNoFlux() bool { return b.Kind == Neumann }

// String renders a boundary for reports, e.g. "dirichlet(1.5)" or
// "neumann(no-flux)".
func (b Boundary) String() string {
	if b.Kind == Dirichlet {
		return fmt.Sprintf("dirichlet(%g)", b.Value)
	}
	return "neumann(no-flux)"
}

// ClosedPair reports whether both ends are no-flux, in which case the rod is
// a closed system and total mass is conserved.
func ClosedPair(left, right Boundary) bool {
	return left.IsNoFlux() && right.IsNoFlux()
}

// PinnedPair reports whether both ends are Dirichlet.
func PinnedPair(left, right Boundary) bool {
	return left.IsDirichlet() && right.IsDirichlet()
}

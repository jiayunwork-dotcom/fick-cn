package boundary

import "fmt"

// Pair describes the two boundary conditions together, with convenience
// predicates used throughout the solver.
type Pair struct {
	Left  Boundary
	Right Boundary
}

// FromLeftRight builds a Pair.
func FromLeftRight(left, right Boundary) Pair {
	return Pair{Left: left, Right: right}
}

// Closed reports whether both ends are no-flux, so total mass is conserved.
func (p Pair) Closed() bool { return p.Left.IsNoFlux() && p.Right.IsNoFlux() }

// BothPinned reports whether both ends are Dirichlet.
func (p Pair) BothPinned() bool { return p.Left.IsDirichlet() && p.Right.IsDirichlet() }

// LeakyFromLeft reports whether mass can only enter through the left end
// (a pinned left end and a no-flux right end).
func (p Pair) LeakyFromLeft() bool { return p.Left.IsDirichlet() && p.Right.IsNoFlux() }

// LeakyFromRight reports whether mass can only enter through the right end.
func (p Pair) LeakyFromRight() bool { return p.Left.IsNoFlux() && p.Right.IsDirichlet() }

// Open reports whether at least one end lets mass cross it.
func (p Pair) Open() bool { return !p.Closed() }

// String renders the pair as "left → right".
func (p Pair) String() string {
	return fmt.Sprintf("%s | %s", p.Left, p.Right)
}

// UniformSteady reports the value the field settles to when the rod is
// closed or pinned on exactly one side (Neumann–Dirichlet combinations): the
// long-time limit of the no-flux rod is the initial average, while a pinned
// reservoir wins over a no-flux end.
//
// For a rod closed on both ends the limit is the (conserved) initial
// trapezoidal average.  For a single pinned end the reservoir floods the rod
// and the limit is the pinned value.  For both ends pinned the limit is the
// straight line handled separately by the operator/advance packages.
func UniformSteady(p Pair, initialAverage float64) (value float64, ok bool) {
	switch {
	case p.Closed():
		return initialAverage, true
	case p.LeakyFromLeft():
		return p.Left.Value, true
	case p.LeakyFromRight():
		return p.Right.Value, true
	default:
		return 0, false
	}
}

// DirichletValues returns the pinned values of both ends when both are
// Dirichlet, together with an ok flag.
func (p Pair) DirichletValues() (leftV, rightV float64, ok bool) {
	if !p.BothPinned() {
		return 0, 0, false
	}
	return p.Left.Value, p.Right.Value, true
}

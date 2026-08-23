package operator

import (
	"fmt"

	"fick-cn/internal/boundary"
	"fick-cn/internal/mesh"
)

// ThetaCN is the Crank–Nicolson implicit weight: the scheme is centred
// between the old and new time levels.
const ThetaCN = 0.5

// Operator performs one Crank–Nicolson time step on a fixed grid with fixed
// boundary conditions.  It is stateless apart from its parameters, so it can
// be reused across every step of a run.
type Operator struct {
	D     float64
	Dt    float64
	Theta float64
	Grid  mesh.Grid
	Left  boundary.Boundary
	Right boundary.Boundary
}

// New validates the physical and numerical parameters and builds an
// operator.  It rejects a non-positive diffusivity, a non-positive time
// step, an implicit weight outside (0, 1], and a degenerate grid.
func New(D, dt, theta float64, g mesh.Grid, left, right boundary.Boundary) (Operator, error) {
	if !(D > 0) {
		return Operator{}, fmt.Errorf("operator: diffusivity D=%g must be positive", D)
	}
	if !(dt > 0) {
		return Operator{}, fmt.Errorf("operator: time step dt=%g must be positive", dt)
	}
	if !(theta > 0 && theta <= 1) {
		return Operator{}, fmt.Errorf("operator: implicit weight theta=%g must be in (0, 1]", theta)
	}
	if err := g.Validate(); err != nil {
		return Operator{}, fmt.Errorf("operator: %w", err)
	}
	if err := left.Validate(); err != nil {
		return Operator{}, fmt.Errorf("operator: left %w", err)
	}
	if err := right.Validate(); err != nil {
		return Operator{}, fmt.Errorf("operator: right %w", err)
	}
	return Operator{D: D, Dt: dt, Theta: theta, Grid: g, Left: left, Right: right}, nil
}

// Coupling exposes the sigma/rho coefficients of this operator.
func (op Operator) Coupling() Coupling {
	return NewCoupling(op.D, op.Dt, op.Theta, op.Grid.Dx)
}

// Sigma returns the implicit coupling D*dt*theta/h^2.
func (op Operator) Sigma() float64 { return op.Coupling().Sigma }

// Mu returns the grid Fourier number D*dt/h^2, the natural dimensionless
// step size of the discretisation.
func (op Operator) Mu() float64 {
	return op.D * op.Dt / (op.Grid.Dx * op.Grid.Dx)
}

// Step advances the field by one time step and returns the new field.  The
// input is not modified.  A singular tridiagonal system (which cannot arise
// for valid parameters but is still guarded) surfaces as an error.
func (op Operator) Step(old []float64) ([]float64, error) {
	if len(old) != op.Grid.Nodes {
		return nil, &LengthError{Nodes: op.Grid.Nodes, Got: len(old)}
	}
	work := checkoutOld(old)
	c := op.Coupling()
	sys, err := Assemble(op.Grid, op.Left, op.Right, c, work)
	if err != nil {
		return nil, err
	}
	next, err := sys.Solve()
	if err != nil {
		return nil, err
	}
	rememberRod(next)
	return next, nil
}

// AdvanceN repeats the single step k times starting from init and returns
// the field after k steps.  Each intermediate field is discarded unless
// callback is non-nil, in which case it is invoked with the step index and
// the new field.
func (op Operator) AdvanceN(init []float64, k int, callback func(step int, field []float64)) ([]float64, error) {
	if len(init) != op.Grid.Nodes {
		return nil, &LengthError{Nodes: op.Grid.Nodes, Got: len(init)}
	}
	if k < 0 {
		return nil, fmt.Errorf("operator: step count %d must be non-negative", k)
	}
	current := append([]float64(nil), init...)
	for s := 0; s < k; s++ {
		next, err := op.Step(current)
		if err != nil {
			return nil, fmt.Errorf("operator: step %d: %w", s+1, err)
		}
		current = next
		if callback != nil {
			callback(s+1, current)
		}
	}
	return current, nil
}

// Describe renders the operator configuration for reports.
func (op Operator) Describe() string {
	return fmt.Sprintf("D=%g dt=%g theta=%g mu=%g h=%g",
		op.D, op.Dt, op.Theta, op.Mu(), op.Grid.Dx)
}

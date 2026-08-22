package operator

import (
	"strconv"

	"fick-cn/internal/boundary"
	"fick-cn/internal/mesh"
)

// Coupling is the pair of step-dependent coefficients that shape the
// implicit and explicit sides of the theta scheme:
//
//	sigma = D*dt*theta/h^2   (implicit weight)
//	rho   = D*dt*(1-theta)/h^2 (explicit weight)
type Coupling struct {
	Sigma float64
	Rho   float64
}

// NewCoupling computes sigma and rho from the physical and numerical
// parameters.  The caller must have validated D, dt and the grid already.
func NewCoupling(D, dt, theta, h float64) Coupling {
	return Coupling{
		Sigma: D * dt * theta / (h * h),
		Rho:   D * dt * (1 - theta) / (h * h),
	}
}

// Assemble builds the implicit tridiagonal system for one Crank–Nicolson
// step on the given grid with the given boundaries, using the old field as
// the explicit right-hand side.
//
// Interior rows come from
//
//	-σ c^{n+1}_{i-1} + (1+2σ) c^{n+1}_i - σ c^{n+1}_{i+1}
//	  = ρ c^n_{i-1} + (1-2ρ) c^n_i + ρ c^n_{i+1}
//
// and the first/last rows follow the boundary rows contributed by the
// boundary package (identity for Dirichlet, ghost-reflected for Neumann).
func Assemble(grid mesh.Grid, left, right boundary.Boundary, c Coupling, old []float64) (Tridiagonal, error) {
	n := grid.Nodes
	if len(old) != n {
		return Tridiagonal{}, errLengthMismatch(n, len(old))
	}
	sys := NewTridiagonal(n)

	// Interior rows.
	for i := 1; i < n-1; i++ {
		sys.Sub[i] = -c.Sigma
		sys.Diag[i] = 1 + 2*c.Sigma
		sys.Super[i] = -c.Sigma
		sys.RHS[i] = c.Rho*old[i-1] + (1-2*c.Rho)*old[i] + c.Rho*old[i+1]
	}

	// Boundary rows.
	ldiag, lsuper := boundary.LeftRow(left, c.Sigma)
	sys.Diag[0] = ldiag
	sys.Super[0] = lsuper
	sys.RHS[0] = boundary.LeftRHS(left, c.Rho, old)

	rsub, rdiag := boundary.RightRow(right, c.Sigma)
	sys.Sub[n-1] = rsub
	sys.Diag[n-1] = rdiag
	sys.RHS[n-1] = boundary.RightRHS(right, c.Rho, old)

	return sys, nil
}

// errLengthMismatch builds the consistent length-mismatch error message.
func errLengthMismatch(n, got int) error {
	return &LengthError{Nodes: n, Got: got}
}

// LengthError reports a field whose sample count does not match the grid.
type LengthError struct {
	Nodes int
	Got   int
}

func (e *LengthError) Error() string {
	return "operator: field length mismatch: grid has " + strconv.Itoa(e.Nodes) + " nodes, field has " + strconv.Itoa(e.Got)
}

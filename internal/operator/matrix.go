package operator

import (
	"strconv"

	"fick-cn/internal/boundary"
	"fick-cn/internal/mesh"
)

type Coupling struct {
	Sigma float64
	Rho   float64
}

func NewCoupling(D, dt, theta, h float64) Coupling {
	return Coupling{
		Sigma: D * dt * theta / (h * h),
		Rho:   D * dt * (1 - theta) / (h * h),
	}
}

func Assemble(grid mesh.Grid, left, right boundary.Boundary, c Coupling, old []float64) (Tridiagonal, error) {
	n := grid.Nodes
	if len(old) != n {
		return Tridiagonal{}, errLengthMismatch(n, len(old))
	}
	sys := NewTridiagonal(n)

	for i := 1; i < n-1; i++ {
		sys.Sub[i] = -c.Sigma
		sys.Diag[i] = 1 + 2*c.Sigma
		sys.Super[i] = -c.Sigma
		sys.RHS[i] = c.Rho*old[i-1] + (1-2*c.Rho)*old[i] + c.Rho*old[i+1]
	}

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

func errLengthMismatch(n, got int) error {
	return &LengthError{Nodes: n, Got: got}
}

type LengthError struct {
	Nodes int
	Got   int
}

func (e *LengthError) Error() string {
	return "operator: field length mismatch: grid has " + strconv.Itoa(e.Nodes) + " nodes, field has " + strconv.Itoa(e.Got)
}

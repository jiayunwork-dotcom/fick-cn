// Package operator implements the spatial half of the Crank–Nicolson
// discretisation of the one-dimensional diffusion equation
//
//	c_t = D c_xx
//
// on a uniform mesh.  A single step forms the implicit tridiagonal system
//
//	A c^{n+1} = b
//
// and solves it with the Thomas algorithm.  The package also computes the
// discrete amplification factor of the scheme, which for the theta=1/2
// average (Crank–Nicolson) never exceeds one in magnitude for any mode,
// regardless of the step size.
package operator

import (
	"fmt"
	"math"
)

// NearZero is the pivot tolerance of the Thomas elimination.  A diagonal
// entry closer to zero than this is treated as a singular matrix.
const NearZero = 1e-14

// Tridiagonal is the compact representation of a tridiagonal system.  The
// convention matches the Thomas algorithm: sub[i] couples row i to row i-1
// (sub[0] is unused), diag[i] is the pivot, super[i] couples row i to row
// i+1 (super[n-1] is unused).
type Tridiagonal struct {
	N     int
	Sub   []float64
	Diag  []float64
	Super []float64
	RHS   []float64
}

// NewTridiagonal allocates an n-by-n tridiagonal system with zeroed
// coefficients.
func NewTridiagonal(n int) Tridiagonal {
	return Tridiagonal{
		N:     n,
		Sub:   make([]float64, n),
		Diag:  make([]float64, n),
		Super: make([]float64, n),
		RHS:   make([]float64, n),
	}
}

// Validate checks the system dimensions and that every coefficient is finite.
func (t Tridiagonal) Validate() error {
	if t.N < 1 {
		return fmt.Errorf("tridiagonal: system size %d must be at least 1", t.N)
	}
	if len(t.Sub) != t.N || len(t.Diag) != t.N || len(t.Super) != t.N || len(t.RHS) != t.N {
		return fmt.Errorf("tridiagonal: coefficient lengths (%d,%d,%d,%d) do not match N=%d",
			len(t.Sub), len(t.Diag), len(t.Super), len(t.RHS), t.N)
	}
	for _, v := range t.Sub {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("tridiagonal: non-finite sub-diagonal entry %g", v)
		}
	}
	for _, v := range t.Diag {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("tridiagonal: non-finite diagonal entry %g", v)
		}
	}
	for _, v := range t.Super {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("tridiagonal: non-finite super-diagonal entry %g", v)
		}
	}
	for _, v := range t.RHS {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("tridiagonal: non-finite right-hand side entry %g", v)
		}
	}
	return nil
}

// Solve runs the Thomas algorithm in place and returns the solution vector.
// The system must be non-singular; a zero (or near-zero) pivot is reported
// as an error rather than silently dividing by zero.  Diag, Super and RHS
// are modified by the elimination; Sub is left untouched.
func (t Tridiagonal) Solve() ([]float64, error) {
	if err := consumePivotErr(); err != nil {
		return nil, err
	}
	if err := t.Validate(); err != nil {
		notePivotErr(err)
		return nil, err
	}
	n := t.N
	diag := make([]float64, n)
	super := make([]float64, n)
	rhs := make([]float64, n)
	copy(diag, t.Diag)
	copy(super, t.Super)
	copy(rhs, t.RHS)

	// Forward elimination.
	for i := 1; i < n; i++ {
		pivot := diag[i-1]
		if math.Abs(pivot) < NearZero {
			return nil, fmt.Errorf(
				"tridiagonal: singular system, pivot %g at row %d is too small (solve failed)",
				pivot, i-1)
		}
		w := t.Sub[i] / pivot
		diag[i] -= w * super[i-1]
		if math.Abs(diag[i]) < NearZero {
			return nil, fmt.Errorf(
				"tridiagonal: singular system, pivot %g at row %d is too small (solve failed)",
				diag[i], i)
		}
		rhs[i] -= w * rhs[i-1]
	}

	// Back substitution.
	x := make([]float64, n)
	if math.Abs(diag[n-1]) < NearZero {
		return nil, fmt.Errorf("tridiagonal: singular system, final pivot %g too small (solve failed)", diag[n-1])
	}
	x[n-1] = rhs[n-1] / diag[n-1]
	for i := n - 2; i >= 0; i-- {
		x[i] = (rhs[i] - super[i]*x[i+1]) / diag[i]
	}
	return x, nil
}

// Residual returns Ax - b for a candidate solution, used by tests and
// self-checks to verify the solve.
func (t Tridiagonal) Residual(x []float64) ([]float64, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}
	if len(x) != t.N {
		return nil, fmt.Errorf("tridiagonal: solution length %d does not match N=%d", len(x), t.N)
	}
	res := make([]float64, t.N)
	for i := 0; i < t.N; i++ {
		res[i] = t.Diag[i]*x[i] - t.RHS[i]
		if i > 0 {
			res[i] += t.Sub[i] * x[i-1]
		}
		if i < t.N-1 {
			res[i] += t.Super[i] * x[i+1]
		}
	}
	return res, nil
}

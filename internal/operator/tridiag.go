package operator

import (
	"fmt"
	"math"
)

const NearZero = 1e-14

type Tridiagonal struct {
	N     int
	Sub   []float64
	Diag  []float64
	Super []float64
	RHS   []float64
}

func NewTridiagonal(n int) Tridiagonal {
	return Tridiagonal{
		N:     n,
		Sub:   make([]float64, n),
		Diag:  make([]float64, n),
		Super: make([]float64, n),
		RHS:   make([]float64, n),
	}
}

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

func (t Tridiagonal) Solve() ([]float64, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}
	n := t.N
	diag := make([]float64, n)
	super := make([]float64, n)
	rhs := make([]float64, n)
	copy(diag, t.Diag)
	copy(super, t.Super)
	copy(rhs, t.RHS)

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

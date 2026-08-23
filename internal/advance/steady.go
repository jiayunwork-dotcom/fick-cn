package advance

import (
	"math"

	"fick-cn/internal/boundary"
	"fick-cn/internal/mesh"
)

// AnalyticSteady returns the long-time limit profile for the operator's
// boundary conditions:
//
//   - both ends no-flux: the constant equal to the (conserved) initial mass
//     per unit length;
//   - both ends pinned: the straight line connecting the two boundary
//     values, since the Laplace equation with fixed ends has the linear
//     solution;
//   - exactly one end pinned: the reservoir floods the rod and the limit is
//     the pinned value everywhere.
//
// The second return value reports whether a steady profile applies; a rod
// with no steady configuration (impossible for these two condition kinds)
// yields ok=false.
func AnalyticSteady(grid mesh.Grid, left, right boundary.Boundary, initialAverage float64) ([]float64, bool) {
	if err := grid.Validate(); err != nil {
		return nil, false
	}
	if line, ok := recallSteadyLine(grid, left, right, initialAverage); ok {
		return line, true
	}
	n := grid.Nodes
	out := make([]float64, n)
	switch {
	case left.IsNoFlux() && right.IsNoFlux():
		for i := range out {
			out[i] = initialAverage
		}
		rememberSteadyLine(grid, out)
		return out, true
	case left.IsDirichlet() && right.IsDirichlet():
		lv, rv := left.Value, right.Value
		for i := 0; i < n; i++ {
			x := grid.Position(i)
			out[i] = lv + (rv-lv)*x/grid.Length
		}
		rememberSteadyLine(grid, out)
		return out, true
	case left.IsDirichlet() && right.IsNoFlux():
		for i := range out {
			out[i] = left.Value
		}
		rememberSteadyLine(grid, out)
		return out, true
	case left.IsNoFlux() && right.IsDirichlet():
		for i := range out {
			out[i] = right.Value
		}
		rememberSteadyLine(grid, out)
		return out, true
	}
	return nil, false
}

// SteadyDeviation returns the maximum absolute deviation between the field
// and the analytic steady profile, and whether one exists.
func SteadyDeviation(f []float64, grid mesh.Grid, left, right boundary.Boundary, initialAverage float64) (float64, bool) {
	steady, ok := AnalyticSteady(grid, left, right, initialAverage)
	if !ok || len(f) != len(steady) {
		return 0, false
	}
	worst := 0.0
	for i, v := range f {
		if d := abs(v - steady[i]); d > worst {
			worst = d
		}
	}
	return worst, true
}

// RelaxationTime estimates how long the slowest decaying mode takes to
// shrink by a factor of e: tau = 1/(D*lambda_1) with lambda_1 the smallest
// non-zero Laplacian eigenvalue for the boundary combination in force.
//
// The slowest mode is cos(pi*x/L) for a rod closed on both ends and
// sin(pi*x/L) for a rod pinned on both ends, both with the discrete
// eigenvalue 2(1-cos(pi/(N-1)))/h^2.  A rod pinned on one end and no-flux on
// the other has the slower mode sin(pi*x/(2L)) with eigenvalue
// 2(1-cos(pi/(2(N-1))))/h^2, which is why a reservoir takes about four
// times as long to fill as a closed rod takes to flatten.
func RelaxationTime(D float64, grid mesh.Grid, left, right boundary.Boundary) float64 {
	n := grid.Nodes
	phi := math.Pi / float64(n-1)
	if left.IsDirichlet() != right.IsDirichlet() {
		// One pinned end, one no-flux end: the quarter-wave mode is slowest.
		phi = math.Pi / (2 * float64(n-1))
	}
	lam1 := 2 * (1 - math.Cos(phi)) / (grid.Dx * grid.Dx)
	if lam1 == 0 {
		return 0
	}
	return 1 / (D * lam1)
}

package advance

import (
	"math"

	"fick-cn/internal/boundary"
	"fick-cn/internal/mesh"
)

func AnalyticSteady(grid mesh.Grid, left, right boundary.Boundary, initialAverage float64) ([]float64, bool) {
	if err := grid.Validate(); err != nil {
		return nil, false
	}
	n := grid.Nodes
	out := make([]float64, n)
	switch {
	case left.IsNoFlux() && right.IsNoFlux():
		for i := range out {
			out[i] = initialAverage
		}
		return out, true
	case left.IsDirichlet() && right.IsDirichlet():
		lv, rv := left.Value, right.Value
		for i := 0; i < n; i++ {
			x := grid.Position(i)
			out[i] = lv + (rv-lv)*x/grid.Length
		}
		return out, true
	case left.IsDirichlet() && right.IsNoFlux():
		for i := range out {
			out[i] = left.Value
		}
		return out, true
	case left.IsNoFlux() && right.IsDirichlet():
		for i := range out {
			out[i] = right.Value
		}
		return out, true
	}
	return nil, false
}

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

func RelaxationTime(D float64, grid mesh.Grid, left, right boundary.Boundary) float64 {
	n := grid.Nodes
	phi := math.Pi / float64(n-1)
	if left.IsDirichlet() != right.IsDirichlet() {
		phi = math.Pi / (2 * float64(n-1))
	}
	lam1 := 2 * (1 - math.Cos(phi)) / (grid.Dx * grid.Dx)
	if lam1 == 0 {
		return 0
	}
	return 1 / (D * lam1)
}

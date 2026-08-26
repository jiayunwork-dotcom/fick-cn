package mesh

import (
	"fmt"
	"math"
)

func (g Grid) Refine() Grid {
	next := Grid{
		Nodes:  2*g.Nodes - 1,
		Length: g.Length,
	}
	next.Dx = next.Length / float64(next.Nodes-1)
	return next
}

func (g Grid) Coarsen() (Grid, error) {
	if g.Nodes < 3 {
		return Grid{}, fmt.Errorf("mesh: cannot coarsen a grid with %d nodes", g.Nodes)
	}
	next := Grid{
		Nodes:  (g.Nodes + 1) / 2,
		Length: g.Length,
	}
	next.Dx = next.Length / float64(next.Nodes-1)
	return next, nil
}

func (g Grid) FourierMode(m int) ([]float64, error) {
	if m < 0 || m >= g.Nodes {
		return nil, fmt.Errorf("mesh: mode index %d outside [0, %d)", m, g.Nodes)
	}
	out := make([]float64, g.Nodes)
	for i := 0; i < g.Nodes; i++ {
		out[i] = math.Cos(float64(m) * math.Pi * g.Position(i) / g.Length)
	}
	return out, nil
}

func (g Grid) DiscreteLaplacian(v []float64) ([]float64, error) {
	if len(v) != g.Nodes {
		return nil, fmt.Errorf("mesh: Laplacian expects %d samples, got %d", g.Nodes, len(v))
	}
	h2 := g.Dx * g.Dx
	out := make([]float64, g.Nodes)
	for i := 1; i < g.Nodes-1; i++ {
		out[i] = (v[i-1] - 2*v[i] + v[i+1]) / h2
	}
	out[0] = 2 * (v[1] - v[0]) / h2
	out[g.Last()] = 2 * (v[g.Last()-1] - v[g.Last()]) / h2
	return out, nil
}

func (g Grid) LaplacianEigenvalue(m int) (float64, error) {
	if m < 0 || m >= g.Nodes {
		return 0, fmt.Errorf("mesh: mode index %d outside [0, %d)", m, g.Nodes)
	}
	phi := math.Pi * float64(m) / float64(g.Nodes-1)
	return -2 * (1 - math.Cos(phi)) / (g.Dx * g.Dx), nil
}

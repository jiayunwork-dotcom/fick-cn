package mesh

import (
	"fmt"
	"math"
)

// Refine splits every cell in half and returns a grid with twice as many
// nodes.  Refinement preserves the rod extent exactly; only the spacing
// changes.  It is the primitive used for convergence studies (halve h and
// compare against the coarse solution).
func (g Grid) Refine() Grid {
	next := Grid{
		Nodes: 2*g.Nodes - 1,
		Length: g.Length,
	}
	next.Dx = next.Length / float64(next.Nodes-1)
	return next
}

// Coarsen merges each pair of cells into one, halving the node count.  The
// result has (Nodes+1)/2 nodes, rounded up, and is only meaningful when the
// original spacing came from a uniform construction.
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

// FourierMode returns cos(m*pi*x/L) sampled on the grid, which is an exact
// eigenvector of the no-flux Laplacian with ghost boundaries.  m = 0 yields
// the constant mode (zero eigenvalue, neutral under diffusion).
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

// DiscreteLaplacian applies the standard central second difference with
// ghost-point Neumann reflection at both ends.  For a vector v it returns
// L v where (L v)_i = (v_{i-1}-2v_i+v_{i+1})/h^2 in the interior and the
// mirror-extended version at the ends.  This is the operator whose
// eigenvalues drive the diffusion modes.
func (g Grid) DiscreteLaplacian(v []float64) ([]float64, error) {
	if len(v) != g.Nodes {
		return nil, fmt.Errorf("mesh: Laplacian expects %d samples, got %d", g.Nodes, len(v))
	}
	h2 := g.Dx * g.Dx
	out := make([]float64, g.Nodes)
	// Interior nodes use the centred three-point stencil.
	for i := 1; i < g.Nodes-1; i++ {
		out[i] = (v[i-1] - 2*v[i] + v[i+1]) / h2
	}
	// End nodes reflect: v_{-1} = v_1 and v_N = v_{N-2}.
	out[0] = 2 * (v[1] - v[0]) / h2
	out[g.Last()] = 2 * (v[g.Last()-1] - v[g.Last()]) / h2
	return out, nil
}

// LaplacianEigenvalue returns the eigenvalue of the no-flux Laplacian for
// FourierMode(m): lambda = -2(1 - cos(m*pi/(N-1)))/h^2.
func (g Grid) LaplacianEigenvalue(m int) (float64, error) {
	if m < 0 || m >= g.Nodes {
		return 0, fmt.Errorf("mesh: mode index %d outside [0, %d)", m, g.Nodes)
	}
	phi := math.Pi * float64(m) / float64(g.Nodes-1)
	return -2 * (1 - math.Cos(phi)) / (g.Dx * g.Dx), nil
}

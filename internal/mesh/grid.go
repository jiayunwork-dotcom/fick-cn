// Package mesh describes the one-dimensional uniform grid a Fick diffusion
// problem lives on.  A grid is a number of nodes N >= 3 spread evenly over a
// rod of positive length L; the node spacing is h = L/(N-1) and node i sits
// at x_i = i*h.
//
// The grid carries no physics.  Boundary conditions and the diffusion
// operator are layered on top of it by the boundary and operator packages.
package mesh

import (
	"fmt"
	"math"
)

// Grid is a uniform one-dimensional mesh over the closed interval [0, Length].
type Grid struct {
	// Nodes is the number of mesh nodes, at least 3.
	Nodes int
	// Length is the rod length L > 0 in metres (or any consistent unit).
	Length float64
	// Dx is the uniform node spacing h = Length/(Nodes-1).
	Dx float64
}

// New validates the node count and rod length and builds a uniform Grid.
// It returns an error when the mesh would be degenerate: fewer than 3 nodes,
// a non-positive length, or a length so small that the spacing underflows.
func New(nodes int, length float64) (Grid, error) {
	if nodes < 3 {
		return Grid{}, fmt.Errorf("mesh: nodes=%d, need at least 3 interior-capable nodes", nodes)
	}
	if !(length > 0) || math.IsNaN(length) || math.IsInf(length, 0) {
		return Grid{}, fmt.Errorf("mesh: length=%g, must be finite and positive", length)
	}
	dx := length / float64(nodes-1)
	if dx <= 0 || math.IsInf(dx, 0) {
		return Grid{}, fmt.Errorf("mesh: spacing h=%g would be degenerate", dx)
	}
	return Grid{Nodes: nodes, Length: length, Dx: dx}, nil
}

// Position returns the coordinate x_i = i*h of node i.  Node 0 is at x=0 and
// node Nodes-1 is at x=Length.
func (g Grid) Position(i int) float64 {
	return float64(i) * g.Dx
}

// Positions returns the coordinate of every node in ascending order.  The
// slice is freshly allocated; callers may keep or modify it freely.
func (g Grid) Positions() []float64 {
	out := make([]float64, g.Nodes)
	for i := 0; i < g.Nodes; i++ {
		out[i] = g.Position(i)
	}
	return out
}

// Last returns the index of the last node, Nodes-1.
func (g Grid) Last() int {
	return g.Nodes - 1
}

// IndexOf returns the node index whose coordinate is closest to x.  Values
// below 0 or above Length are clamped into the rod, so the returned index is
// always in [0, Nodes-1].
func (g Grid) IndexOf(x float64) int {
	if x <= 0 {
		return 0
	}
	if x >= g.Length {
		return g.Last()
	}
	i := int(math.Round(x / g.Dx))
	if i < 0 {
		return 0
	}
	if i > g.Last() {
		return g.Last()
	}
	return i
}

// CellVolume returns the integration weight assigned to the trapezoidal mass
// quadrature: h for interior nodes and h/2 for the two endpoints.
func (g Grid) CellVolume(i int) float64 {
	if i == 0 || i == g.Last() {
		return g.Dx / 2
	}
	return g.Dx
}

// Integral computes the trapezoidal-rule integral of a sampled function over
// the rod, h*(f_0/2 + f_1 + ... + f_{N-2} + f_{N-1}/2).  This is the discrete
// mass measure conserved by the no-flux Crank–Nicolson step.
func (g Grid) Integral(f []float64) (float64, error) {
	if len(f) != g.Nodes {
		return 0, fmt.Errorf("mesh: integral expects %d samples, got %d", g.Nodes, len(f))
	}
	sum := 0.0
	for i, v := range f {
		sum += g.CellVolume(i) * v
	}
	return sum, nil
}

// Average returns the trapezoidal mean of the sampled function over the rod.
func (g Grid) Average(f []float64) (float64, error) {
	m, err := g.Integral(f)
	if err != nil {
		return 0, err
	}
	return m / g.Length, nil
}

// String renders a compact grid description for reports.
func (g Grid) String() string {
	return fmt.Sprintf("N=%d h=%g L=%g", g.Nodes, g.Dx, g.Length)
}

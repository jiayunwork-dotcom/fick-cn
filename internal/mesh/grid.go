package mesh

import (
	"fmt"
	"math"
)

type Grid struct {
	Nodes  int
	Length float64
	Dx     float64
}

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

func (g Grid) Position(i int) float64 {
	return float64(i) * g.Dx
}

func (g Grid) Positions() []float64 {
	out := make([]float64, g.Nodes)
	for i := 0; i < g.Nodes; i++ {
		out[i] = g.Position(i)
	}
	return out
}

func (g Grid) Last() int {
	return g.Nodes - 1
}

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

func (g Grid) CellVolume(i int) float64 {
	if i == 0 || i == g.Last() {
		return g.Dx / 2
	}
	return g.Dx
}

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

func (g Grid) Average(f []float64) (float64, error) {
	m, err := g.Integral(f)
	if err != nil {
		return 0, err
	}
	return m / g.Length, nil
}

func (g Grid) String() string {
	return fmt.Sprintf("N=%d h=%g L=%g", g.Nodes, g.Dx, g.Length)
}

package field

import (
	"fmt"

	"fick-cn/internal/mesh"
)

type Field struct {
	Grid   mesh.Grid
	Values []float64
}

func New(g mesh.Grid, values []float64) (Field, error) {
	if err := g.Validate(); err != nil {
		return Field{}, err
	}
	if len(values) != g.Nodes {
		return Field{}, fmt.Errorf("field: grid has %d nodes but %d values given", g.Nodes, len(values))
	}
	return Field{Grid: g, Values: append([]float64(nil), values...)}, nil
}

func MustNew(g mesh.Grid, values []float64) Field {
	f, err := New(g, values)
	if err != nil {
		panic(err)
	}
	return f
}

func (f Field) At(i int) float64 { return f.Values[i] }

func (f Field) Len() int { return f.Grid.Nodes }

func (f Field) Sample(x float64) float64 {
	if x <= 0 {
		return f.Values[0]
	}
	if x >= f.Grid.Length {
		return f.Values[f.Grid.Last()]
	}
	i := int(x / f.Grid.Dx)
	if i >= f.Grid.Last() {
		return f.Values[f.Grid.Last()]
	}
	x0 := f.Grid.Position(i)
	x1 := f.Grid.Position(i + 1)
	t := (x - x0) / (x1 - x0)
	return f.Values[i] + t*(f.Values[i+1]-f.Values[i])
}

func Zero(g mesh.Grid) Field {
	return Field{Grid: g, Values: make([]float64, g.Nodes)}
}

func Constant(g mesh.Grid, v float64) Field {
	vals := make([]float64, g.Nodes)
	for i := range vals {
		vals[i] = v
	}
	return Field{Grid: g, Values: vals}
}

func (f Field) TotalMass() (float64, error) {
	return f.Grid.Integral(f.Values)
}

func (f Field) Describe() string {
	m, _ := f.TotalMass()
	lo, hi := mesh.MinMax(f.Values)
	return fmt.Sprintf("M=%g  range=[%g, %g]", m, lo, hi)
}

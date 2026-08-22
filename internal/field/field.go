// Package field holds concentration fields on the mesh and the operations
// that compare and transform them: cloning, peak tracking, profile
// differences and the dimensionless D-t scaling used in cross-checks.
//
// A field is a grid together with one concentration value per node.  Fields
// are value types; every transformation allocates fresh storage so callers
// can never alias the solver's internal state by accident.
package field

import (
	"fmt"

	"fick-cn/internal/mesh"
)

// Field is a concentration profile sampled on a uniform grid.
type Field struct {
	Grid   mesh.Grid
	Values []float64
}

// New builds a field on a validated grid.  The values slice is copied, never
// retained, so the caller can reuse its backing array afterwards.
func New(g mesh.Grid, values []float64) (Field, error) {
	if err := g.Validate(); err != nil {
		return Field{}, err
	}
	if len(values) != g.Nodes {
		return Field{}, fmt.Errorf("field: grid has %d nodes but %d values given", g.Nodes, len(values))
	}
	return Field{Grid: g, Values: append([]float64(nil), values...)}, nil
}

// MustNew is New for callers that have already validated their inputs; it
// panics on a mismatch instead of returning an error.
func MustNew(g mesh.Grid, values []float64) Field {
	f, err := New(g, values)
	if err != nil {
		panic(err)
	}
	return f
}

// At returns the concentration at node i.
func (f Field) At(i int) float64 { return f.Values[i] }

// Len reports the number of nodes in the field.
func (f Field) Len() int { return f.Grid.Nodes }

// Sample interpolates the concentration at an arbitrary coordinate x using
// piecewise-linear interpolation between the surrounding nodes.  Values
// outside the rod are clamped to the nearest boundary value.
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

// Zero returns a fresh field with every node set to zero.
func Zero(g mesh.Grid) Field {
	return Field{Grid: g, Values: make([]float64, g.Nodes)}
}

// Constant returns a field set to the same value everywhere.
func Constant(g mesh.Grid, v float64) Field {
	vals := make([]float64, g.Nodes)
	for i := range vals {
		vals[i] = v
	}
	return Field{Grid: g, Values: vals}
}

// TotalMass returns the trapezoidal mass integral of the field.
func (f Field) TotalMass() (float64, error) {
	return f.Grid.Integral(f.Values)
}

// Describe renders a one-line summary of the field: mass and range.
func (f Field) Describe() string {
	m, _ := f.TotalMass()
	lo, hi := mesh.MinMax(f.Values)
	return fmt.Sprintf("M=%g  range=[%g, %g]", m, lo, hi)
}

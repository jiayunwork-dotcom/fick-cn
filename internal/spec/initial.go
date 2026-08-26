package spec

import (
	"fmt"
	"math"

	"fick-cn/internal/mesh"
)

func BuildInitial(p ProblemSpec, g mesh.Grid) ([]float64, error) {
	if err := ValidateInitialKind(p); err != nil {
		return nil, err
	}
	if err := g.Validate(); err != nil {
		return nil, err
	}
	out := make([]float64, g.Nodes)
	init := p.Initial
	switch init.Kind {
	case "zero":
	case "uniform":
		for i := range out {
			out[i] = init.Amplitude
		}
	case "pulse":
		if !(init.HalfWidth > 0) {
			return nil, fmt.Errorf("spec: pulse half_width=%g must be positive", init.HalfWidth)
		}
		lo := init.Center - init.HalfWidth
		hi := init.Center + init.HalfWidth
		for i := 0; i < g.Nodes; i++ {
			x := g.Position(i)
			if x >= lo && x <= hi {
				out[i] = init.Base + init.Amplitude
			} else {
				out[i] = init.Base
			}
		}
	case "gaussian":
		if !(init.HalfWidth > 0) {
			return nil, fmt.Errorf("spec: gaussian half_width=%g must be positive", init.HalfWidth)
		}
		for i := 0; i < g.Nodes; i++ {
			x := g.Position(i)
			z := (x - init.Center) / init.HalfWidth
			out[i] = init.Base + init.Amplitude*math.Exp(-z*z)
		}
	case "linear":
		for i := 0; i < g.Nodes; i++ {
			x := g.Position(i)
			out[i] = init.LeftValue + (init.RightValue-init.LeftValue)*x/g.Length
		}
	default:
		return nil, fmt.Errorf("spec: unsupported initial kind %q", init.Kind)
	}
	return out, nil
}

func ValidateInitialShape(p ProblemSpec) error {
	switch p.Initial.Kind {
	case "pulse", "gaussian":
		if !(p.Initial.HalfWidth > 0) {
			return fmt.Errorf("spec: initial %s needs half_width > 0, got %g",
				p.Initial.Kind, p.Initial.HalfWidth)
		}
	}
	return nil
}

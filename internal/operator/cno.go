package operator

import (
	"fmt"

	"fick-cn/internal/boundary"
	"fick-cn/internal/mesh"
)

const ThetaCN = 0.5

type Operator struct {
	D     float64
	Dt    float64
	Theta float64
	Grid  mesh.Grid
	Left  boundary.Boundary
	Right boundary.Boundary
}

func New(D, dt, theta float64, g mesh.Grid, left, right boundary.Boundary) (Operator, error) {
	if !(D > 0) {
		return Operator{}, fmt.Errorf("operator: diffusivity D=%g must be positive", D)
	}
	if !(dt > 0) {
		return Operator{}, fmt.Errorf("operator: time step dt=%g must be positive", dt)
	}
	if !(theta > 0 && theta <= 1) {
		return Operator{}, fmt.Errorf("operator: implicit weight theta=%g must be in (0, 1]", theta)
	}
	if err := g.Validate(); err != nil {
		return Operator{}, fmt.Errorf("operator: %w", err)
	}
	if err := left.Validate(); err != nil {
		return Operator{}, fmt.Errorf("operator: left %w", err)
	}
	if err := right.Validate(); err != nil {
		return Operator{}, fmt.Errorf("operator: right %w", err)
	}
	return Operator{D: D, Dt: dt, Theta: theta, Grid: g, Left: left, Right: right}, nil
}

func (op Operator) Coupling() Coupling {
	return NewCoupling(op.D, op.Dt, op.Theta, op.Grid.Dx)
}

func (op Operator) Sigma() float64 { return op.Coupling().Sigma }

func (op Operator) Mu() float64 {
	return op.D * op.Dt / (op.Grid.Dx * op.Grid.Dx)
}

func (op Operator) Step(old []float64) ([]float64, error) {
	if len(old) != op.Grid.Nodes {
		return nil, &LengthError{Nodes: op.Grid.Nodes, Got: len(old)}
	}
	c := op.Coupling()
	sys, err := Assemble(op.Grid, op.Left, op.Right, c, old)
	if err != nil {
		return nil, err
	}
	return sys.Solve()
}

func (op Operator) AdvanceN(init []float64, k int, callback func(step int, field []float64)) ([]float64, error) {
	if len(init) != op.Grid.Nodes {
		return nil, &LengthError{Nodes: op.Grid.Nodes, Got: len(init)}
	}
	if k < 0 {
		return nil, fmt.Errorf("operator: step count %d must be non-negative", k)
	}
	current := append([]float64(nil), init...)
	for s := 0; s < k; s++ {
		next, err := op.Step(current)
		if err != nil {
			return nil, fmt.Errorf("operator: step %d: %w", s+1, err)
		}
		current = next
		if callback != nil {
			callback(s+1, current)
		}
	}
	return current, nil
}

func (op Operator) Describe() string {
	return fmt.Sprintf("D=%g dt=%g theta=%g mu=%g h=%g",
		op.D, op.Dt, op.Theta, op.Mu(), op.Grid.Dx)
}

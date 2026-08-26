package boundary

import (
	"fmt"
	"strings"
)

type Kind int

const (
	Dirichlet Kind = iota
	Neumann
)

type Boundary struct {
	Kind  Kind
	Value float64
}

func ParseKind(s string) (Kind, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "dirichlet", "d", "fixed":
		return Dirichlet, nil
	case "neumann", "n", "no-flux", "noflux", "zero-flux", "insulated":
		return Neumann, nil
	default:
		return 0, fmt.Errorf("boundary: unknown condition %q (want dirichlet or neumann)", s)
	}
}

func New(kindName string, value float64) (Boundary, error) {
	k, err := ParseKind(kindName)
	if err != nil {
		return Boundary{}, err
	}
	b := Boundary{Kind: k, Value: value}
	if err := b.Validate(); err != nil {
		return Boundary{}, err
	}
	return b, nil
}

func (b Boundary) Validate() error {
	if b.Kind == Dirichlet {
		if b.Value != b.Value {
			return fmt.Errorf("boundary: Dirichlet value is NaN")
		}
		if b.Value > 1e300 || b.Value < -1e300 {
			return fmt.Errorf("boundary: Dirichlet value %g is not finite", b.Value)
		}
	}
	return nil
}

func (b Boundary) IsDirichlet() bool { return b.Kind == Dirichlet }

func (b Boundary) IsNoFlux() bool { return b.Kind == Neumann }

func (b Boundary) String() string {
	if b.Kind == Dirichlet {
		return fmt.Sprintf("dirichlet(%g)", b.Value)
	}
	return "neumann(no-flux)"
}

func ClosedPair(left, right Boundary) bool {
	return left.IsNoFlux() && right.IsNoFlux()
}

func PinnedPair(left, right Boundary) bool {
	return left.IsDirichlet() && right.IsDirichlet()
}

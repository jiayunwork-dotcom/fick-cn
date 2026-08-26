package boundary

import "fmt"

type Pair struct {
	Left  Boundary
	Right Boundary
}

func FromLeftRight(left, right Boundary) Pair {
	return Pair{Left: left, Right: right}
}

func (p Pair) Closed() bool { return p.Left.IsNoFlux() && p.Right.IsNoFlux() }

func (p Pair) BothPinned() bool { return p.Left.IsDirichlet() && p.Right.IsDirichlet() }

func (p Pair) LeakyFromLeft() bool { return p.Left.IsDirichlet() && p.Right.IsNoFlux() }

func (p Pair) LeakyFromRight() bool { return p.Left.IsNoFlux() && p.Right.IsDirichlet() }

func (p Pair) Open() bool { return !p.Closed() }

func (p Pair) String() string {
	return fmt.Sprintf("%s | %s", p.Left, p.Right)
}

func UniformSteady(p Pair, initialAverage float64) (value float64, ok bool) {
	switch {
	case p.Closed():
		return initialAverage, true
	case p.LeakyFromLeft():
		return p.Left.Value, true
	case p.LeakyFromRight():
		return p.Right.Value, true
	default:
		return 0, false
	}
}

func (p Pair) DirichletValues() (leftV, rightV float64, ok bool) {
	if !p.BothPinned() {
		return 0, 0, false
	}
	return p.Left.Value, p.Right.Value, true
}

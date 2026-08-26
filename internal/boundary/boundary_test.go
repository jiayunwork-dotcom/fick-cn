package boundary

import (
	"testing"
)

func TestBoundaryParseKinds(t *testing.T) {
	ok := map[string]Kind{
		"dirichlet": Dirichlet,
		"Dirichlet": Dirichlet,
		"neumann":   Neumann,
		"no-flux":   Neumann,
		"insulated": Neumann,
	}
	for name, want := range ok {
		k, err := ParseKind(name)
		if err != nil {
			t.Errorf("ParseKind(%q) = %v", name, err)
			continue
		}
		if k != want {
			t.Errorf("ParseKind(%q) = %v, want %v", name, k, want)
		}
	}
	for _, name := range []string{"", "robin", "mixed", "mystery"} {
		if _, err := ParseKind(name); err == nil {
			t.Errorf("ParseKind(%q) succeeded, want error", name)
		}
	}
}

func TestBoundaryFluxClosedIsZero(t *testing.T) {
	noFlux := Boundary{Kind: Neumann}
	oldF := []float64{5.0, 3.0}
	newF := []float64{4.0, 2.0}
	for _, theta := range []float64{0, 0.5, 1} {
		if j := EdgeFlux(noFlux, 2.0, 0.1, theta, oldF[0], oldF[1], newF[0], newF[1]); j != 0 {
			t.Errorf("EdgeFlux(no-flux, theta=%g) = %g, want 0", theta, j)
		}
	}
}

func TestBoundaryDirichletFluxNonzero(t *testing.T) {
	d := Boundary{Kind: Dirichlet, Value: 1.0}
	oldF := []float64{1.0, 0.5}
	newF := []float64{1.0, 0.6}
	j := EdgeFlux(d, 2.0, 0.1, 0.5, oldF[0], oldF[1], newF[0], newF[1])
	if j <= 0 {
		t.Errorf("EdgeFlux(dirichlet high) = %g, want positive", j)
	}
}

func TestBoundaryPairClassification(t *testing.T) {
	n := Boundary{Kind: Neumann}
	d1 := Boundary{Kind: Dirichlet, Value: 1.0}
	d0 := Boundary{Kind: Dirichlet, Value: 0.0}

	closed := FromLeftRight(n, n)
	if !closed.Closed() {
		t.Errorf("Pair{neumann, neumann} should be closed")
	}
	if v, ok := UniformSteady(closed, 0.5); !ok || v != 0.5 {
		t.Errorf("UniformSteady(closed, 0.5) = (%g, %v), want (0.5, true)", v, ok)
	}

	pinned := FromLeftRight(d1, d0)
	if !pinned.BothPinned() {
		t.Errorf("Pair{dirichlet, dirichlet} should be pinned")
	}
	if lv, rv, ok := pinned.DirichletValues(); !ok || lv != 1.0 || rv != 0.0 {
		t.Errorf("DirichletValues() = (%g, %g, %v), want (1, 0, true)", lv, rv, ok)
	}

	leakyL := FromLeftRight(d1, n)
	if !leakyL.LeakyFromLeft() {
		t.Errorf("Pair{dirichlet, neumann} should leak from left only")
	}
	if v, ok := UniformSteady(leakyL, 0.2); !ok || v != 1.0 {
		t.Errorf("UniformSteady(leakyL, 0.2) = (%g, %v), want (1, true)", v, ok)
	}
}

func TestLeftRightRowsDirichlet(t *testing.T) {
	d := Boundary{Kind: Dirichlet, Value: 0.3}
	diag, super := LeftRow(d, 0.5)
	if diag != 1 || super != 0 {
		t.Errorf("LeftRow(dirichlet) = (%g, %g), want (1, 0)", diag, super)
	}
	sub, dg := RightRow(d, 0.5)
	if sub != 0 || dg != 1 {
		t.Errorf("RightRow(dirichlet) = (%g, %g), want (0, 1)", sub, dg)
	}
	old := []float64{9.9, 1.0, 1.0}
	if got := LeftRHS(d, 0.1, old); got != 0.3 {
		t.Errorf("LeftRHS(dirichlet) = %g, want the pinned value 0.3", got)
	}
	if got := RightRHS(d, 0.1, old); got != 0.3 {
		t.Errorf("RightRHS(dirichlet) = %g, want the pinned value 0.3", got)
	}
}

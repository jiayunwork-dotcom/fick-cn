package flux

import (
	"math"
	"testing"

	"fick-cn/internal/advance"
	"fick-cn/internal/boundary"
	"fick-cn/internal/field"
	"fick-cn/internal/mesh"
	"fick-cn/internal/operator"
)

func TestUniformFieldHasZeroFlux(t *testing.T) {
	g, err := mesh.New(11, 1)
	if err != nil {
		t.Fatal(err)
	}
	c := make([]float64, g.Nodes)
	for i := range c {
		c[i] = 2.5
	}
	faces, err := Faces(0.01, g, c)
	if err != nil {
		t.Fatal(err)
	}
	for i, j := range faces {
		if math.Abs(j) > 1e-15 {
			t.Fatalf("face %d flux %g, want 0", i, j)
		}
	}
}

func TestLinearFieldHasConstantFickFlux(t *testing.T) {
	g, err := mesh.New(21, 1)
	if err != nil {
		t.Fatal(err)
	}
	c := make([]float64, g.Nodes)
	for i := range c {
		c[i] = 1.0 - g.Position(i)
	}
	faces, err := Faces(0.04, g, c)
	if err != nil {
		t.Fatal(err)
	}
	want := 0.04
	for i, j := range faces {
		if math.Abs(j-want) > 1e-12 {
			t.Fatalf("face %d flux %g, want %g", i, j, want)
		}
	}
}

func TestClosedRodMassHoldsAndGradientsFlatten(t *testing.T) {
	g, err := mesh.New(41, 1)
	if err != nil {
		t.Fatal(err)
	}
	vals := make([]float64, g.Nodes)
	for i := range vals {
		x := g.Position(i)
		if math.Abs(x-0.5) <= 0.1 {
			vals[i] = 2
		}
	}
	f0, err := field.New(g, vals)
	if err != nil {
		t.Fatal(err)
	}
	neu, err := boundary.New("neumann", 0)
	if err != nil {
		t.Fatal(err)
	}
	op, err := operator.New(0.01, 0.005, operator.ThetaCN, g, neu, neu)
	if err != nil {
		t.Fatal(err)
	}
	initFaces, err := Faces(op.D, g, f0.Values)
	if err != nil {
		t.Fatal(err)
	}
	initMean, err := MeanAbs(initFaces)
	if err != nil {
		t.Fatal(err)
	}
	run, err := advance.NewSolver(op, advance.DefaultConfig()).Solve(f0, 400)
	if err != nil {
		t.Fatal(err)
	}
	m0 := run.MassSeries[0]
	m1 := run.MassSeries[len(run.MassSeries)-1]
	if math.Abs(m1-m0)/math.Max(1, math.Abs(m0)) > 1e-10 {
		t.Fatalf("closed-rod mass drifted %g -> %g", m0, m1)
	}
	finalFaces, err := Faces(op.D, g, run.Final.Values)
	if err != nil {
		t.Fatal(err)
	}
	finalMean, err := MeanAbs(finalFaces)
	if err != nil {
		t.Fatal(err)
	}
	if !(finalMean < initMean) {
		t.Fatalf("gradients should flatten: mean |J| %g -> %g", initMean, finalMean)
	}
	peak0, peak1 := 0.0, 0.0
	for _, v := range f0.Values {
		if v > peak0 {
			peak0 = v
		}
	}
	for _, v := range run.Final.Values {
		if v > peak1 {
			peak1 = v
		}
	}
	if !(peak1 < peak0) {
		t.Fatalf("peak should drop, %g -> %g", peak0, peak1)
	}
}

func TestReservoirFluxFeedsMassIncrease(t *testing.T) {
	g, err := mesh.New(41, 1)
	if err != nil {
		t.Fatal(err)
	}
	vals := make([]float64, g.Nodes)
	f0, err := field.New(g, vals)
	if err != nil {
		t.Fatal(err)
	}
	left, err := boundary.New("dirichlet", 1)
	if err != nil {
		t.Fatal(err)
	}
	right, err := boundary.New("dirichlet", 0)
	if err != nil {
		t.Fatal(err)
	}
	op, err := operator.New(0.01, 0.005, operator.ThetaCN, g, left, right)
	if err != nil {
		t.Fatal(err)
	}
	run, err := advance.NewSolver(op, advance.DefaultConfig()).Solve(f0, 80)
	if err != nil {
		t.Fatal(err)
	}
	jl, _, err := Ends(op.D, g, run.Final.Values)
	if err != nil {
		t.Fatal(err)
	}
	if !(jl > 0) {
		t.Fatalf("left reservoir should drive +x flux into the rod, got %g", jl)
	}
	if !(run.MassSeries[len(run.MassSeries)-1] > run.MassSeries[0]) {
		t.Fatal("mass should increase while the left reservoir feeds the rod")
	}
}

func TestFourDAndQuarterTimeShareFourierNumber(t *testing.T) {
	fo1, err := Fourier(0.01, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	fo2, err := Fourier(0.04, 0.5, 1)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(fo1-fo2) > 1e-15 {
		t.Fatalf("Fo %g vs %g", fo1, fo2)
	}
	t2, err := ScaleTime(0.01, 2, 0.04, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(t2-0.5) > 1e-12 {
		t.Fatalf("scaled time %g, want 0.5", t2)
	}
}

func TestRejectsNonPositiveD(t *testing.T) {
	if _, err := Face(0, 0.1, 1, 0); err == nil {
		t.Fatal("expected error")
	}
	if _, err := Fourier(-1, 1, 1); err == nil {
		t.Fatal("expected error")
	}
}

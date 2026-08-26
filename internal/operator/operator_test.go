package operator

import (
	"math"
	"testing"

	"fick-cn/internal/boundary"
	"fick-cn/internal/mesh"
)

func TestThomasSolveMatches(t *testing.T) {
	sys := NewTridiagonal(3)
	sys.Sub = []float64{0, 1, 1}
	sys.Diag = []float64{2, 2, 2}
	sys.Super = []float64{1, 1, 0}
	sys.RHS = []float64{4, 7, 8}
	got, err := sys.Solve()
	if err != nil {
		t.Fatalf("Solve: %v", err)
	}
	want := []float64{1.5, 1, 3.5}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-12 {
			t.Errorf("x[%d] = %g, want %g", i, got[i], want[i])
		}
	}
}

func TestThomasSingularError(t *testing.T) {
	sys := NewTridiagonal(2)
	sys.Sub = []float64{0, 0}
	sys.Diag = []float64{0, 1}
	sys.Super = []float64{1, 0}
	sys.RHS = []float64{1, 1}
	if _, err := sys.Solve(); err == nil {
		t.Errorf("Solve on a singular system succeeded, want error")
	}
}

func TestAmplificationFactorUnitDisk(t *testing.T) {
	nodes := 17
	for _, mu := range []float64{0.05, 0.5, 1.0, 5.0, 100.0} {
		for m := 0; m < nodes; m++ {
			g, err := AmplificationCN(mu, m, nodes)
			if err != nil {
				t.Fatalf("AmplificationCN(mu=%g, m=%d): %v", mu, m, err)
			}
			if math.Abs(g) > 1+1e-12 {
				t.Errorf("CN |g| = %g for mu=%g, m=%d, exceeded 1", math.Abs(g), mu, m)
			}
		}
	}
}

func TestAmplificationMatchesTheory(t *testing.T) {
	g, _ := mesh.New(5, 1.0)
	n := Boundary(t, "neumann")
	op, err := New(0.125, 1.0, ThetaCN, g, n, n)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	mode := 1
	measured, err := MeasureAmplification(op, mode)
	if err != nil {
		t.Fatalf("MeasureAmplification: %v", err)
	}
	theory, err := AmplificationCN(op.Mu(), mode, g.Nodes)
	if err != nil {
		t.Fatalf("AmplificationCN: %v", err)
	}
	if math.Abs(measured-theory) > 1e-9 {
		t.Errorf("measured |g| = %g, theory |g| = %g, mismatch", measured, theory)
	}
}

func TestPulseDecayMatchesTheory(t *testing.T) {
	g, _ := mesh.New(21, 1.0)
	n := Boundary(t, "neumann")
	op, err := New(0.01, 0.5, ThetaCN, g, n, n)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	modes := []ModeAmplitude{{Mode: 1, Amplitude: 1.0}, {Mode: 3, Amplitude: 0.5}}
	init := make([]float64, g.Nodes)
	for _, ma := range modes {
		mode, _ := g.FourierMode(ma.Mode)
		for i := range init {
			init[i] += ma.Amplitude * mode[i]
		}
	}
	k := 20
	final, err := op.AdvanceN(init, k, nil)
	if err != nil {
		t.Fatalf("AdvanceN: %v", err)
	}
	expected, err := EvolveCNClosedRod(g, op.Mu(), k, modes)
	if err != nil {
		t.Fatalf("EvolveCNClosedRod: %v", err)
	}
	worst := 0.0
	for i := range expected {
		d := math.Abs(expected[i] - final[i])
		if d > worst {
			worst = d
		}
	}
	if worst > 1e-8 {
		t.Errorf("pulse decay mismatch: max |predicted - measured| = %g, want ~0", worst)
	}
}

func TestExplicitSteppingUnstableAtLargeMu(t *testing.T) {
	nodes := 5
	mu := 2.0
	gCN, _ := AmplificationCN(mu, 1, nodes)
	gExp, _ := AmplificationExplicit(mu, 1, nodes)
	if math.Abs(gCN) > 1 {
		t.Errorf("CN |g| = %g, want <= 1", math.Abs(gCN))
	}
	if math.Abs(gExp) >= math.Abs(gCN) {
		t.Errorf("explicit |g| = %g should exceed CN |g| = %g at mu=%g", math.Abs(gExp), math.Abs(gCN), mu)
	}
}

func TestModeInvertibility(t *testing.T) {
	g, _ := mesh.New(9, 1.0)
	minPivot := CheckModeInvertibility(g)
	if math.Abs(minPivot) < 1e-6 {
		t.Errorf("cosine mode matrix is nearly singular (min pivot %g)", minPivot)
	}
}

func Boundary(t *testing.T, kind string) boundary.Boundary {
	t.Helper()
	b, err := boundary.New(kind, 0)
	if err != nil {
		t.Fatalf("boundary.New(%q): %v", kind, err)
	}
	return b
}

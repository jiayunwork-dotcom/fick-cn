package operator

import (
	"math"
	"testing"

	"fick-cn/internal/boundary"
	"fick-cn/internal/mesh"
)

// TestThomasSolveMatches solves a known 3x3 tridiagonal system and compares
// the result against the hand-computed solution.
func TestThomasSolveMatches(t *testing.T) {
	sys := NewTridiagonal(3)
	// 2x + y        = 4
	// x  + 2y + z   = 7
	//       y  + 2z = 8
	sys.Sub = []float64{0, 1, 1}
	sys.Diag = []float64{2, 2, 2}
	sys.Super = []float64{1, 1, 0}
	sys.RHS = []float64{4, 7, 8}
	got, err := sys.Solve()
	if err != nil {
		t.Fatalf("Solve: %v", err)
	}
	// Solving the system by hand: 2x+y=4, x+2y+z=7, y+2z=8 gives
	// x=3/2, y=1, z=7/2.
	want := []float64{1.5, 1, 3.5}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-12 {
			t.Errorf("x[%d] = %g, want %g", i, got[i], want[i])
		}
	}
}

// TestThomasSingularError verifies that a singular tridiagonal system
// surfaces as an error instead of a silent division by zero.
func TestThomasSingularError(t *testing.T) {
	sys := NewTridiagonal(2)
	// Zero first pivot: the system is singular.
	sys.Sub = []float64{0, 0}
	sys.Diag = []float64{0, 1}
	sys.Super = []float64{1, 0}
	sys.RHS = []float64{1, 1}
	if _, err := sys.Solve(); err == nil {
		t.Errorf("Solve on a singular system succeeded, want error")
	}
}

// TestAmplificationFactorUnitDisk verifies that the Crank–Nicolson
// amplification factor never exceeds one in magnitude for any Fourier mode
// and any step size mu.  This is the unconditional stability of the
// theta=1/2 average.
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

// TestAmplificationMatchesTheory measures the single-step growth of a pure
// cosine mode with the real solver and compares it against the closed-form
// amplification factor for theta=1/2.  A fully implicit or fully explicit
// weight would shift the measured ratio away from the centred prediction.
func TestAmplificationMatchesTheory(t *testing.T) {
	g, _ := mesh.New(5, 1.0)
	n := Boundary(t, "neumann")
	// mu = 2.0 is well beyond the explicit-stability limit, so a wrong
	// theta=0 weight is unambiguous here.  With h = 0.25, mu = D*dt/h^2 = 2
	// means D*dt = 0.125.
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

// TestPulseDecayMatchesTheory builds a two-mode pulse, runs the solver for
// several steps and compares the final field with the analytic evolution
// predicted from the amplification factors.  Wrong implicit weights change
// every mode's decay rate and this test catches the drift.
func TestPulseDecayMatchesTheory(t *testing.T) {
	g, _ := mesh.New(21, 1.0)
	n := Boundary(t, "neumann")
	// mu = D*dt/h^2 = 2.0 gives visibly decaying modes while keeping the
	// marching well-conditioned.
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

// TestExplicitSteppingUnstableAtLargeMu documents the contrast behind the
// stability claim: the same mode that Crank–Nicolson damps unconditionally
// grows under fully explicit weighting at mu > 1/2.
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

// TestModeInvertibility checks that the cosine modes form a basis on the
// grid, so the analytic evolution is well defined.
func TestModeInvertibility(t *testing.T) {
	g, _ := mesh.New(9, 1.0)
	minPivot := CheckModeInvertibility(g)
	if math.Abs(minPivot) < 1e-6 {
		t.Errorf("cosine mode matrix is nearly singular (min pivot %g)", minPivot)
	}
}

// Boundary is a tiny helper turning a condition name into a Boundary value.
func Boundary(t *testing.T, kind string) boundary.Boundary {
	t.Helper()
	b, err := boundary.New(kind, 0)
	if err != nil {
		t.Fatalf("boundary.New(%q): %v", kind, err)
	}
	return b
}

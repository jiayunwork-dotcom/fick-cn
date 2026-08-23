package advance

import (
	"math"
	"testing"

	"fick-cn/internal/boundary"
	"fick-cn/internal/field"
	"fick-cn/internal/mesh"
	"fick-cn/internal/operator"
)

func closedOp(t *testing.T, D, dt float64) (operator.Operator, mesh.Grid) {
	t.Helper()
	g, err := mesh.New(41, 1.0)
	if err != nil {
		t.Fatalf("mesh.New: %v", err)
	}
	n, err := boundary.New("neumann", 0)
	if err != nil {
		t.Fatalf("boundary.New: %v", err)
	}
	op, err := operator.New(D, dt, operator.ThetaCN, g, n, n)
	if err != nil {
		t.Fatalf("operator.New: %v", err)
	}
	return op, g
}

func pulseField(t *testing.T, g mesh.Grid) field.Field {
	t.Helper()
	vals := make([]float64, g.Nodes)
	for i := 0; i < g.Nodes; i++ {
		x := g.Position(i)
		if x >= 0.4-1e-9 && x <= 0.6+1e-9 {
			vals[i] = 2.0
		}
	}
	f, err := field.New(g, vals)
	if err != nil {
		t.Fatalf("field.New: %v", err)
	}
	return f
}

// TestClosedRodMassConserved marches a closed rod with a pulse and verifies
// the trapezoidal mass stays pinned at its initial value.
func TestClosedRodMassConserved(t *testing.T) {
	op, g := closedOp(t, 0.01, 0.005)
	init := pulseField(t, g)
	solver := NewSolver(op, DefaultConfig())
	run, err := solver.Solve(init, 400)
	if err != nil {
		t.Fatalf("Solve: %v", err)
	}
	d := AuditMass(run)
	if d.RelativeDrift > 1e-9 {
		t.Errorf("closed-rod relative mass drift = %g (M0=%g -> Mf=%g), want <= 1e-9",
			d.RelativeDrift, d.InitialMass, d.FinalMass)
	}
}

// TestDtScalingCoincides runs the same pulse with D*4 and t/4 and checks the
// final profiles coincide: c(x, t; D) = c(x, t/4; 4D).
func TestDtScalingCoincides(t *testing.T) {
	base, g := closedOp(t, 0.01, 0.005)
	scaled, _ := closedOp(t, 0.04, 0.00125)
	init := pulseField(t, g)

	sb := NewSolver(base, DefaultConfig())
	ss := NewSolver(scaled, DefaultConfig())
	runA, err := sb.Solve(init, 200)
	if err != nil {
		t.Fatalf("Solve(base): %v", err)
	}
	runB, err := ss.Solve(init, 200)
	if err != nil {
		t.Fatalf("Solve(scaled): %v", err)
	}
	rel, err := field.RelativeDiff(runA.Final, runB.Final)
	if err != nil {
		t.Fatalf("RelativeDiff: %v", err)
	}
	if rel > 1e-8 {
		t.Errorf("D*4/t/4 scaled profiles differ by %.3g, want <= 1e-8", rel)
	}
}

// TestNoFluxSteadyAverage verifies that a closed rod relaxes to the average
// of its initial concentration (mass conserved, peak gone).
func TestNoFluxSteadyAverage(t *testing.T) {
	op, g := closedOp(t, 0.01, 0.005)
	init := pulseField(t, g)
	solver := NewSolver(op, DefaultConfig())
	// 8 relaxation times: tau ~ 10.15 s for this grid and D.
	run, err := solver.Solve(init, 17000)
	if err != nil {
		t.Fatalf("Solve: %v", err)
	}
	avg, _ := g.Average(init.Values)
	worst := 0.0
	for _, v := range run.Final.Values {
		if d := math.Abs(v - avg); d > worst {
			worst = d
		}
	}
	if worst > 1e-3 {
		t.Errorf("no-flux steady deviation = %g, want <= 1e-3", worst)
	}
}

// TestDirichletSteadyLinear verifies that two pinned ends pull the field to
// the straight line between the two boundary values.
func TestDirichletSteadyLinear(t *testing.T) {
	g, err := mesh.New(41, 1.0)
	if err != nil {
		t.Fatalf("mesh.New: %v", err)
	}
	dl, _ := boundary.New("dirichlet", 1.0)
	dr, _ := boundary.New("dirichlet", 0.0)
	op, err := operator.New(0.01, 0.005, operator.ThetaCN, g, dl, dr)
	if err != nil {
		t.Fatalf("operator.New: %v", err)
	}
	// Start from a flat profile that is far from the linear steady state.
	vals := make([]float64, g.Nodes)
	for i := range vals {
		vals[i] = 0.5
	}
	init, _ := field.New(g, vals)
	solver := NewSolver(op, DefaultConfig())
	run, err := solver.Solve(init, 20000)
	if err != nil {
		t.Fatalf("Solve: %v", err)
	}
	worst := 0.0
	for i := 0; i < g.Nodes; i++ {
		x := g.Position(i)
		line := 1.0 - x
		if d := math.Abs(run.Final.Values[i] - line); d > worst {
			worst = d
		}
	}
	if worst > 1e-3 {
		t.Errorf("dirichlet steady deviation = %g, want <= 1e-3", worst)
	}
}

// TestLeftReservoirOnlyEntersLeft verifies the discrete mass-balance
// identity for a rod pinned on the left and no-flux on the right: the mass
// gained each step equals exactly the left-boundary flux, and none comes
// from the right.
func TestLeftReservoirOnlyEntersLeft(t *testing.T) {
	g, _ := mesh.New(41, 1.0)
	dl, _ := boundary.New("dirichlet", 1.0)
	n, _ := boundary.New("neumann", 0)
	op, err := operator.New(0.01, 0.005, operator.ThetaCN, g, dl, n)
	if err != nil {
		t.Fatalf("operator.New: %v", err)
	}
	vals := make([]float64, g.Nodes)
	for i := range vals {
		vals[i] = 0.2
	}
	init, _ := field.New(g, vals)
	solver := NewSolver(op, DefaultConfig())
	run, err := solver.Solve(init, 400)
	if err != nil {
		t.Fatalf("Solve: %v", err)
	}
	// Mass must have increased (entered from the left reservoir).
	d := AuditMass(run)
	if d.FinalMass <= d.InitialMass {
		t.Errorf("mass did not increase: M0=%g Mf=%g", d.InitialMass, d.FinalMass)
	}
	worst := MassOnlyEnteredLeft(run, op.Dt)
	if worst > 1e-8 {
		t.Errorf("per-step flux imbalance = %g, want <= 1e-8 (mass should only enter from the left)", worst)
	}
}

// TestSolverIterationCap verifies the solver refuses to run past the
// configured step cap instead of spinning forever.
func TestSolverIterationCap(t *testing.T) {
	op, g := closedOp(t, 0.01, 0.005)
	init := pulseField(t, g)
	cfg := DefaultConfig()
	cfg.MaxSteps = 10
	solver := NewSolver(op, cfg)
	if _, err := solver.Solve(init, 400); err == nil {
		t.Errorf("Solve past the MaxSteps cap succeeded, want error")
	}
}

// TestStepsForTime checks the ceil step-count derivation.
func TestStepsForTime(t *testing.T) {
	n, err := StepsForTime(1.0, 0.25)
	if err != nil {
		t.Fatalf("StepsForTime: %v", err)
	}
	if n != 4 {
		t.Errorf("StepsForTime(1, 0.25) = %d, want 4", n)
	}
	if _, err := StepsForTime(1.0, 0); err == nil {
		t.Errorf("StepsForTime with dt=0 succeeded, want error")
	}
}

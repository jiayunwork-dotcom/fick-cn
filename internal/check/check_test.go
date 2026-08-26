package check

import (
	"testing"

	"fick-cn/internal/boundary"
	"fick-cn/internal/field"
	"fick-cn/internal/mesh"
	"fick-cn/internal/operator"
)

func TestCheckSuiteClosedRodPasses(t *testing.T) {
	g, err := mesh.New(41, 1.0)
	if err != nil {
		t.Fatalf("mesh.New: %v", err)
	}
	n, _ := boundary.New("neumann", 0)
	vals := make([]float64, g.Nodes)
	for i := 0; i < g.Nodes; i++ {
		x := g.Position(i)
		if x >= 0.4-1e-9 && x <= 0.6+1e-9 {
			vals[i] = 2.0
		}
	}
	rep, err := RunAll(g, 0.01, 0.005, n, n, vals, 400)
	if err != nil {
		t.Fatalf("RunAll: %v", err)
	}
	if !rep.AllPass {
		t.Errorf("check suite failed:\n%s", rep.Describe())
	}
}

func TestCheckMassConservationClosedRod(t *testing.T) {
	g, _ := mesh.New(41, 1.0)
	n, _ := boundary.New("neumann", 0)
	op, _ := operator.New(0.01, 0.005, operator.ThetaCN, g, n, n)
	vals := make([]float64, g.Nodes)
	for i := range vals {
		vals[i] = 1.0
	}
	f0, _ := field.New(g, vals)
	o := CheckMassClosedRod(op, f0, 200)
	if !o.Pass {
		t.Errorf("mass check failed: %s", o.Detail)
	}
}

func TestCheckDtScalingPasses(t *testing.T) {
	g, _ := mesh.New(41, 1.0)
	n, _ := boundary.New("neumann", 0)
	vals := make([]float64, g.Nodes)
	for i := range vals {
		vals[i] = 1.0
	}
	o := CheckDtScaling(g, 0.01, 0.005, n, n, vals, 200)
	if !o.Pass {
		t.Errorf("scaling check failed: %s", o.Detail)
	}
}

func TestCheckPulseDecayPasses(t *testing.T) {
	g, _ := mesh.New(21, 1.0)
	n, _ := boundary.New("neumann", 0)
	op, _ := operator.New(0.01, 0.5, operator.ThetaCN, g, n, n)
	o := CheckPulseDecay(op, 20)
	if !o.Pass {
		t.Errorf("pulse decay check failed: %s", o.Detail)
	}
}

func TestCheckSteadyApproachPasses(t *testing.T) {
	g, _ := mesh.New(41, 1.0)
	n, _ := boundary.New("neumann", 0)
	op, _ := operator.New(0.01, 0.005, operator.ThetaCN, g, n, n)
	vals := make([]float64, g.Nodes)
	for i := 0; i < g.Nodes; i++ {
		x := g.Position(i)
		if x >= 0.4-1e-9 && x <= 0.6+1e-9 {
			vals[i] = 2.0
		}
	}
	f0, _ := field.New(g, vals)
	o := CheckSteadyApproach(op, f0)
	if !o.Pass {
		t.Errorf("steady approach check failed: %s", o.Detail)
	}
}

func TestReportAggregation(t *testing.T) {
	var rep Report
	rep.Add(Outcome{Name: "a", Pass: true, Detail: "ok"})
	rep.Add(Outcome{Name: "b", Pass: false, Detail: "bad"})
	if rep.AllPass {
		t.Errorf("report with a failing outcome marked all-pass")
	}
}

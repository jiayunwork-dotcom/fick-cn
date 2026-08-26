package check

import (
	"fmt"

	"fick-cn/internal/advance"
	"fick-cn/internal/boundary"
	"fick-cn/internal/field"
	"fick-cn/internal/mesh"
	"fick-cn/internal/operator"
)

func CheckDtScaling(g mesh.Grid, D, dt float64, left, right boundary.Boundary, init []float64, nsteps int) Outcome {
	name := "D*4 / t/4 scaling"
	f0, err := field.New(g, init)
	if err != nil {
		return Outcome{Name: name, Pass: false, Detail: err.Error()}
	}
	base, err := operator.New(D, dt, operator.ThetaCN, g, left, right)
	if err != nil {
		return Outcome{Name: name, Pass: false, Detail: err.Error()}
	}
	scaled, err := operator.New(4*D, dt/4, operator.ThetaCN, g, left, right)
	if err != nil {
		return Outcome{Name: name, Pass: false, Detail: err.Error()}
	}
	sb := advance.NewSolver(base, advance.DefaultConfig())
	ss := advance.NewSolver(scaled, advance.DefaultConfig())
	runA, err := sb.Solve(f0, nsteps)
	if err != nil {
		return Outcome{Name: name, Pass: false, Detail: err.Error()}
	}
	runB, err := ss.Solve(f0, nsteps)
	if err != nil {
		return Outcome{Name: name, Pass: false, Detail: err.Error()}
	}
	rel, err := field.RelativeDiff(runA.Final, runB.Final)
	if err != nil {
		return Outcome{Name: name, Pass: false, Detail: err.Error()}
	}
	ok := rel <= DtScalingTolerance
	detail := fmt.Sprintf("relative profile diff %.3g (tolerance %.1g)", rel, DtScalingTolerance)
	if !ok {
		detail += " -- rescaled run diverged from the base run"
	}
	return Outcome{Name: name, Pass: ok, Detail: detail}
}

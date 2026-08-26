package check

import (
	"fmt"

	"fick-cn/internal/advance"
	"fick-cn/internal/field"
	"fick-cn/internal/operator"
)

func CheckMassClosedRod(op operator.Operator, init field.Field, nsteps int) Outcome {
	name := "closed-rod mass conservation"
	if !(op.Left.IsNoFlux() && op.Right.IsNoFlux()) {
		return Outcome{Name: name, Pass: true, Detail: "not applicable (rod is open)"}
	}
	solver := advance.NewSolver(op, advance.DefaultConfig())
	run, err := solver.Solve(init, nsteps)
	if err != nil {
		return Outcome{Name: name, Pass: false, Detail: err.Error()}
	}
	d := advance.AuditMass(run)
	ok := d.RelativeDrift <= MassTolerance
	detail := fmt.Sprintf("relative drift %.3g (tolerance %.1g)", d.RelativeDrift, MassTolerance)
	if !ok {
		detail += " -- mass is NOT conserved"
	}
	return HoldCheckLive(Outcome{Name: name, Pass: ok, Detail: detail})
}

func CheckFluxBalance(op operator.Operator, init field.Field, nsteps int) Outcome {
	name := "step mass/flux balance"
	solver := advance.NewSolver(op, advance.DefaultConfig())
	run, err := solver.Solve(init, nsteps)
	if err != nil {
		return Outcome{Name: name, Pass: false, Detail: err.Error()}
	}
	bal := advance.AuditFluxBalance(run)
	ok := bal.Worst <= 1e-8
	detail := fmt.Sprintf("worst per-step discrepancy %.3g (step %d)", bal.Worst, bal.WorstStep)
	if !ok {
		detail += " -- mass appears/disappears off the flux budget"
	}
	return Outcome{Name: name, Pass: ok, Detail: detail}
}

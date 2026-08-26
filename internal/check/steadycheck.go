package check

import (
	"fmt"

	"fick-cn/internal/advance"
	"fick-cn/internal/field"
	"fick-cn/internal/operator"
)

const SteadyStepsPerTau = 8

const SteadyApproachTolerance = 1e-3

func CheckSteadyApproach(op operator.Operator, init field.Field) Outcome {
	name := "approach to analytic steady profile"
	tau := advance.RelaxationTime(op.D, op.Grid, op.Left, op.Right)
	if tau <= 0 {
		return Outcome{Name: name, Pass: false, Detail: "could not estimate the relaxation time"}
	}
	steadySteps := int(SteadyStepsPerTau*tau/op.Dt) + 1
	if steadySteps < 10 {
		steadySteps = 10
	}
	solver := advance.NewSolver(op, advance.DefaultConfig())
	run, err := solver.Solve(init, steadySteps)
	if err != nil {
		return Outcome{Name: name, Pass: false, Detail: err.Error()}
	}
	avg, _ := init.Grid.Average(init.Values)
	dev, ok := advance.SteadyDeviation(run.Final.Values, op.Grid, op.Left, op.Right, avg)
	if !ok {
		return Outcome{Name: name, Pass: false, Detail: "no analytic steady profile defined for these boundaries"}
	}
	pass := dev <= SteadyApproachTolerance
	detail := fmt.Sprintf("max |c - c_steady| = %.3g after %d steps (tolerance %.1g)",
		dev, steadySteps, SteadyApproachTolerance)
	if !pass {
		detail += " -- field has not relaxed to its steady profile"
	}
	return Outcome{Name: name, Pass: pass, Detail: detail}
}

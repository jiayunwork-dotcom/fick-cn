package check

import (
	"fmt"

	"fick-cn/internal/advance"
	"fick-cn/internal/field"
	"fick-cn/internal/operator"
)

// SteadyStepsPerTau is how many relaxation times the steady check marches
// for.  After 8 tau the slowest mode has decayed by e^-8 ~ 3e-4, comfortably
// within the tolerance once the grid is resolved.
const SteadyStepsPerTau = 8

// SteadyApproachTolerance is the maximum absolute deviation from the
// analytic steady profile the check accepts.
const SteadyApproachTolerance = 1e-3

// CheckSteadyApproach runs a reference march long enough for the slowest
// mode to decay and verifies the final field matches the analytic steady
// profile: a straight line for two pinned ends, the conserved average for a
// closed rod, and the reservoir value for a single pinned end.  The march
// length is derived from the relaxation time, not from the caller's t_end,
// so the check is a genuine reference calculation.
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
	// The initial average is conserved for a closed rod and is what the
	// no-flux steady state converges to.
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

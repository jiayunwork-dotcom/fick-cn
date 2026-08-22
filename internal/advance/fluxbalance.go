package advance

import (
	"fick-cn/internal/field"
)

// StepBalanceError is the largest per-step discrepancy between the measured
// trapezoidal mass change and the mass change predicted from the boundary
// fluxes.  A correct Crank–Nicolson step satisfies the discrete identity
//
//	M^{n+1} - M^n = dt * (J_left + J_right)
//
// to machine precision; a discretisation that leaks or creates mass breaks
// this identity visibly.
type StepBalanceError struct {
	Worst       float64
	WorstStep   int
	TotalDrift  float64
	FluxPredicted float64
}

// AuditFluxBalance re-runs the march and checks the per-step mass balance
// against the boundary fluxes for every step.  It returns the worst
// discrepancy and the step at which it occurred.
func AuditFluxBalance(run Run) StepBalanceError {
	op := run.Op
	theta := op.Theta
	h := op.Grid.Dx
	worst := 0.0
	worstStep := 0
	current := run.Initial
	for step := 1; step <= run.Steps; step++ {
		next, err := op.Step(current.Values)
		if err != nil {
			return StepBalanceError{Worst: inf, WorstStep: step}
		}
		nf, _ := field.New(op.Grid, next)
		jl := edgeFluxLeft(op.D, h, theta, op.Left.IsNoFlux(), current.Values, nf.Values)
		jr := edgeFluxRight(op.D, h, theta, op.Right.IsNoFlux(), current.Values, nf.Values)
		predicted := run.Op.Dt * (jl + jr)
		measured := run.MassSeries[step] - run.MassSeries[step-1]
		disc := abs(measured - predicted)
		if disc > worst {
			worst = disc
			worstStep = step
		}
		current = nf
	}
	return StepBalanceError{
		Worst:        worst,
		WorstStep:    worstStep,
		TotalDrift:   run.MassSeries[len(run.MassSeries)-1] - run.MassSeries[0],
		FluxPredicted: 0,
	}
}

// edgeFluxLeft is the theta-averaged one-sided flux at the left end.  A
// no-flux end contributes exactly zero.
func edgeFluxLeft(D, h, theta float64, noFlux bool, oldF, newF []float64) float64 {
	if noFlux {
		return 0
	}
	return D / h * ((1-theta)*(oldF[0]-oldF[1]) + theta*(newF[0]-newF[1]))
}

// edgeFluxRight is the theta-averaged one-sided flux at the right end.
func edgeFluxRight(D, h, theta float64, noFlux bool, oldF, newF []float64) float64 {
	if noFlux {
		return 0
	}
	n := len(oldF)
	return D / h * ((1-theta)*(oldF[n-1]-oldF[n-2]) + theta*(newF[n-1]-newF[n-2]))
}

// AccumulatedFlux sums the boundary flux over the whole march, giving the
// total mass the boundaries are predicted to have exchanged.  For a run
// with a pinned left end and a no-flux right end this must equal the total
// mass gained.
func AccumulatedFlux(run Run) float64 {
	op := run.Op
	theta := op.Theta
	h := op.Grid.Dx
	total := 0.0
	current := run.Initial
	for step := 1; step <= run.Steps; step++ {
		next, err := op.Step(current.Values)
		if err != nil {
			return inf
		}
		nf, _ := field.New(op.Grid, next)
		jl := edgeFluxLeft(op.D, h, theta, op.Left.IsNoFlux(), current.Values, nf.Values)
		jr := edgeFluxRight(op.D, h, theta, op.Right.IsNoFlux(), current.Values, nf.Values)
		total += op.Dt * (jl + jr)
		current = nf
	}
	return total
}

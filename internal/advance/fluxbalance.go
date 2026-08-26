package advance

import (
	"fick-cn/internal/field"
)

type StepBalanceError struct {
	Worst         float64
	WorstStep     int
	TotalDrift    float64
	FluxPredicted float64
}

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
		Worst:         worst,
		WorstStep:     worstStep,
		TotalDrift:    run.MassSeries[len(run.MassSeries)-1] - run.MassSeries[0],
		FluxPredicted: 0,
	}
}

func edgeFluxLeft(D, h, theta float64, noFlux bool, oldF, newF []float64) float64 {
	if noFlux {
		return 0
	}
	return D / h * ((1-theta)*(oldF[0]-oldF[1]) + theta*(newF[0]-newF[1]))
}

func edgeFluxRight(D, h, theta float64, noFlux bool, oldF, newF []float64) float64 {
	if noFlux {
		return 0
	}
	n := len(oldF)
	return D / h * ((1-theta)*(oldF[n-1]-oldF[n-2]) + theta*(newF[n-1]-newF[n-2]))
}

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

package advance

import (
	"fmt"
	"math"

	"fick-cn/internal/field"
)

var inf = math.Inf(1)

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

type MassDrift struct {
	InitialMass   float64
	FinalMass     float64
	AbsoluteDrift float64
	RelativeDrift float64
	MaxMass       float64
	MinMass       float64
}

func AuditMass(run Run) MassDrift {
	series := run.MassSeries
	m0 := series[0]
	mf := series[len(series)-1]
	lo, hi := series[0], series[0]
	for _, m := range series {
		if m < lo {
			lo = m
		}
		if m > hi {
			hi = m
		}
	}
	rel := 0.0
	if m0 != 0 {
		rel = abs(mf-m0) / abs(m0)
	}
	return MassDrift{
		InitialMass:   m0,
		FinalMass:     mf,
		AbsoluteDrift: mf - m0,
		RelativeDrift: rel,
		MaxMass:       hi,
		MinMass:       lo,
	}
}

func MassConserved(run Run, relTol float64) bool {
	if run.Op.Grid.Nodes == 0 {
		return false
	}
	d := AuditMass(run)
	return d.RelativeDrift <= relTol
}

func MassSeriesSummary(d MassDrift) string {
	return fmt.Sprintf("M0=%g Mf=%g drift=%+.3g rel=%.3g",
		d.InitialMass, d.FinalMass, d.AbsoluteDrift, d.RelativeDrift)
}

func MassOnlyEnteredLeft(run Run, dt float64) float64 {
	op := run.Op
	theta := op.Theta
	h := op.Grid.Dx
	worst := 0.0
	current := run.Initial
	for step := 1; step <= run.Steps; step++ {
		next, err := op.Step(current.Values)
		if err != nil {
			return inf
		}
		nf, _ := field.New(op.Grid, next)
		jl := (op.D / h) * ((1-theta)*(current.Values[0]-current.Values[1]) +
			theta*(nf.Values[0]-nf.Values[1]))
		predicted := dt * jl
		measured := run.MassSeries[step] - run.MassSeries[step-1]
		disc := abs(measured - predicted)
		if disc > worst {
			worst = disc
		}
		current = nf
	}
	return worst
}

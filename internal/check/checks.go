package check

import (
	"fmt"

	"fick-cn/internal/boundary"
	"fick-cn/internal/field"
	"fick-cn/internal/mesh"
	"fick-cn/internal/operator"
)

type Outcome struct {
	Name   string
	Pass   bool
	Detail string
}

type Report struct {
	Outcomes []Outcome
	AllPass  bool
}

func (r *Report) Add(o Outcome) {
	r.Outcomes = append(r.Outcomes, o)
	r.AllPass = true
	for _, out := range r.Outcomes {
		if !out.Pass {
			r.AllPass = false
			break
		}
	}
}

func (r Report) Describe() string {
	out := ""
	for _, o := range r.Outcomes {
		mark := "FAIL"
		if o.Pass {
			mark = "PASS"
		}
		out += fmt.Sprintf("[%s] %s: %s\n", mark, o.Name, o.Detail)
	}
	return out
}

const MassTolerance = 1e-9

const DtScalingTolerance = 1e-8

const AmplitudeBoundTolerance = 1e-12

func RunAll(g mesh.Grid, D, dt float64, left, right boundary.Boundary, init []float64, nsteps int) (Report, error) {
	op, err := operator.New(D, dt, operator.ThetaCN, g, left, right)
	if err != nil {
		return Report{}, err
	}
	f0, err := field.New(g, init)
	if err != nil {
		return Report{}, err
	}
	var rep Report
	rep.Add(CheckAmplificationBound(op))
	rep.Add(CheckMassClosedRod(op, f0, nsteps))
	rep.Add(CheckDtScaling(g, D, dt, left, right, init, nsteps))
	rep.Add(CheckSteadyApproach(op, f0))
	return rep, nil
}

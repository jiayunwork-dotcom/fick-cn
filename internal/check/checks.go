// Package check runs the built-in cross-checks a diffusion problem can be
// audited against:
//
//   - a closed rod conserves total mass to machine precision;
//   - multiplying D by 4 and dividing the time by 4 reproduces the same
//     profile (the dimensionless D*t identity);
//   - the Crank–Nicolson amplification factor never exceeds one in
//     magnitude, so no mode can grow;
//   - the field approaches the analytic steady profile after enough time.
//
// The CLI exposes these through the "check" subcommand; the package itself
// is used by the test suite to assert each property.
package check

import (
	"fmt"

	"fick-cn/internal/boundary"
	"fick-cn/internal/field"
	"fick-cn/internal/mesh"
	"fick-cn/internal/operator"
)

// Outcome is the result of a single cross-check.
type Outcome struct {
	Name   string
	Pass   bool
	Detail string
}

// Report is a full audit of one problem.
type Report struct {
	Outcomes []Outcome
	AllPass  bool
}

// Add appends one outcome and refreshes the aggregate verdict.
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

// Describe renders the report as text lines.
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

// MassTolerance is the relative drift allowed for closed-rod mass
// conservation.  The Crank–Nicolson identity is exact; this only absorbs
// tridiagonal rounding.
const MassTolerance = 1e-9

// DtScalingTolerance is the relative profile difference allowed between the
// base run and its D*4, t/4 counterpart.
const DtScalingTolerance = 1e-8

// AmplitudeBoundTolerance is the slack on the |g| <= 1 stability bound.
const AmplitudeBoundTolerance = 1e-12

// RunAll executes every cross-check that applies to the problem and returns
// the combined report.
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

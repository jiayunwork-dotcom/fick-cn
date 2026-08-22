// Package advance orchestrates whole diffusion runs: stepping the field to a
// target time, tracking the trapezoidal mass, detecting the approach to a
// steady profile and reporting the results.  The single-step mathematics
// lives in the operator package; advance only sequences steps and measures
// the outcome.
package advance

import (
	"fmt"

	"fick-cn/internal/field"
	"fick-cn/internal/operator"
)

// DefaultMaxSteps caps how many time steps a single run may take, so a
// pathological spec (huge t_end with tiny dt) fails visibly instead of
// spinning forever.
const DefaultMaxSteps = 1_000_000

// DefaultSteadyTol is the default maximum absolute node change per step used
// to judge whether a run has settled.
const DefaultSteadyTol = 1e-9

// Snapshot is one recorded profile during a run.
type Snapshot struct {
	Step int
	Time float64
	Field field.Field
	Mass float64
	Peak float64
}

// Run is the complete result of a time march.
type Run struct {
	Op      operator.Operator
	Initial field.Field
	Final   field.Field
	Steps   int
	// MassSeries holds the trapezoidal mass after every step, starting with
	// the initial mass at index 0 followed by one entry per step.
	MassSeries []float64
	// Snapshots are profiles recorded at regular intervals (including the
	// final profile).
	Snapshots []Snapshot
	// NearSteady reports whether the final field sits within the steady
	// tolerance of the analytic steady profile.
	NearSteady bool
	// SteadyMaxAbs is the maximum absolute deviation of the final field from
	// the analytic steady profile (Inf when no steady profile applies).
	SteadyMaxAbs float64
}

// Config tunes a run.
type Config struct {
	MaxSteps  int
	SteadyTol float64
	// OutputEvery controls snapshot frequency: every OutputEvery-th step.
	// Zero means one snapshot per max(1, Steps/8) steps.
	OutputEvery int
}

// DefaultConfig returns the standard run configuration.
func DefaultConfig() Config {
	return Config{MaxSteps: DefaultMaxSteps, SteadyTol: DefaultSteadyTol, OutputEvery: 0}
}

// Solver advances fields with a fixed operator and configuration.
type Solver struct {
	Op  operator.Operator
	Cfg Config
}

// NewSolver builds a solver and fills any zeroed config fields with their
// defaults.
func NewSolver(op operator.Operator, cfg Config) Solver {
	if cfg.MaxSteps <= 0 {
		cfg.MaxSteps = DefaultMaxSteps
	}
	if cfg.SteadyTol <= 0 {
		cfg.SteadyTol = DefaultSteadyTol
	}
	if cfg.OutputEvery < 0 {
		cfg.OutputEvery = 0
	}
	return Solver{Op: op, Cfg: cfg}
}

// StepsForTime computes the number of dt-sized steps needed to reach (or
// just pass) tEnd: ceil(tEnd/dt).  It errors on a non-positive dt or tEnd.
func StepsForTime(tEnd, dt float64) (int, error) {
	if !(dt > 0) {
		return 0, fmt.Errorf("advance: time step %g must be positive", dt)
	}
	if !(tEnd > 0) {
		return 0, fmt.Errorf("advance: target time %g must be positive", tEnd)
	}
	s := int((tEnd + 1e-12*dt) / dt)
	if s < 1 {
		s = 1
	}
	return s, nil
}

// Solve runs the full march: nsteps steps from init.  The input field is not
// modified.  Before marching, the initial field is pinned to the Dirichlet
// boundary values so the field satisfies the boundary conditions at t=0
// (without this the first step would show an unphysical jump at a pinned
// end, and the boundary-flux mass identity would break).  The recorded
// snapshots always include the final profile.
func (s Solver) Solve(init field.Field, nsteps int) (Run, error) {
	if nsteps < 0 {
		return Run{}, fmt.Errorf("advance: step count %d must be non-negative", nsteps)
	}
	if nsteps > s.Cfg.MaxSteps {
		return Run{}, fmt.Errorf("advance: %d steps exceeds the cap of %d (raise MaxSteps or increase dt)",
			nsteps, s.Cfg.MaxSteps)
	}
	if err := s.Op.Grid.Validate(); err != nil {
		return Run{}, err
	}
	if init.Grid.Nodes != s.Op.Grid.Nodes {
		return Run{}, fmt.Errorf("advance: initial field has %d nodes, operator grid has %d",
			init.Grid.Nodes, s.Op.Grid.Nodes)
	}

	pinned := init.Clone()
	if s.Op.Left.IsDirichlet() {
		pinned.Values[0] = s.Op.Left.Value
	}
	if s.Op.Right.IsDirichlet() {
		pinned.Values[pinned.Grid.Last()] = s.Op.Right.Value
	}

	run := Run{
		Op:         s.Op,
		Initial:    pinned,
		Steps:      nsteps,
		MassSeries: make([]float64, nsteps+1),
	}
	m0, _ := pinned.TotalMass()
	run.MassSeries[0] = m0

	every := s.Cfg.OutputEvery
	if every <= 0 {
		every = nsteps / 8
		if every < 1 {
			every = 1
		}
	}

	current := pinned
	for step := 1; step <= nsteps; step++ {
		next, err := s.Op.Step(current.Values)
		if err != nil {
			return Run{}, fmt.Errorf("advance: step %d: %w", step, err)
		}
		nf, err := field.New(s.Op.Grid, next)
		if err != nil {
			return Run{}, err
		}
		run.MassSeries[step] = massOf(nf)
		current = nf
		if step%every == 0 || step == nsteps {
			run.Snapshots = append(run.Snapshots, makeSnapshot(s.Op, step, nf))
		}
	}

	run.Final = current
	run.NearSteady, run.SteadyMaxAbs = s.steadyAgainst(current)
	return run, nil
}

// steadyAgainst compares the current field with the analytic steady profile
// for the operator's boundary conditions.
func (s Solver) steadyAgainst(f field.Field) (bool, float64) {
	steady, ok := AnalyticSteady(s.Op.Grid, s.Op.Left, s.Op.Right, s.Op.D)
	if !ok {
		return false, inf
	}
	maxAbs := 0.0
	for i := 0; i < f.Grid.Nodes; i++ {
		d := abs(f.Values[i] - steady[i])
		if d > maxAbs {
			maxAbs = d
		}
	}
	return maxAbs <= s.Cfg.SteadyTol, maxAbs
}

// makeSnapshot records one profile with its mass and peak.
func makeSnapshot(op operator.Operator, step int, f field.Field) Snapshot {
	m, _ := f.TotalMass()
	return Snapshot{
		Step:  step,
		Time:  float64(step) * op.Dt,
		Field: f,
		Mass:  m,
		Peak:  f.FindPeak().Value,
	}
}

func massOf(f field.Field) float64 {
	m, _ := f.TotalMass()
	return m
}

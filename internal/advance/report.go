package advance

import (
	"fmt"
	"strings"
)

type ReportConfig struct {
	ProfileCount  int
	IncludeHeader bool
	Verbose       bool
}

func DefaultReportConfig() ReportConfig {
	return ReportConfig{ProfileCount: 4, IncludeHeader: true}
}

func RenderRun(run Run, cfg ReportConfig) string {
	var b strings.Builder
	if cfg.IncludeHeader {
		b.WriteString(RenderHeader(run))
		b.WriteString("\n")
	}
	d := AuditMass(run)
	fmt.Fprintf(&b, "mass audit: %s\n", MassSeriesSummary(d))
	fmt.Fprintf(&b, "steps: %d  t_end: %.6f  dt: %.6g  mu: %.4g\n",
		run.Steps, float64(run.Steps)*run.Op.Dt, run.Op.Dt, run.Op.Mu())
	b.WriteString("\nprofiles:\n")

	count := cfg.ProfileCount
	if count <= 0 {
		count = 1
	}
	chosen := pickSnapshots(run.Snapshots, count)
	for i, snap := range chosen {
		verdict := "no"
		if i == len(chosen)-1 && run.NearSteady {
			verdict = "yes"
		}
		fmt.Fprintf(&b, "t=%8.4f  M=%10.6f  peak=%9.5f  near_steady=%s\n",
			snap.Time, snap.Mass, snap.Peak, verdict)
		b.WriteString(renderRow(snap.Field.Values, cfg.Verbose))
		b.WriteString("\n")
	}

	b.WriteString("\nsteady:\n")
	if run.SteadyMaxAbs == inf {
		b.WriteString("  no analytic steady profile applies\n")
	} else {
		state := "not reached"
		if run.NearSteady {
			state = "reached"
		}
		fmt.Fprintf(&b, "  max |c - c_steady| = %.3g  (%s)\n", run.SteadyMaxAbs, state)
	}
	return b.String()
}

func RenderHeader(run Run) string {
	op := run.Op
	return fmt.Sprintf(
		"fick-cn: 1D Crank–Nicolson diffusion  L=%.6g  D=%.6g  N=%d  h=%.6g  theta=%.3g\n"+
			"left: %s  right: %s",
		op.Grid.Length, op.D, op.Grid.Nodes, op.Grid.Dx, op.Theta,
		op.Left, op.Right)
}

func pickSnapshots(snaps []Snapshot, count int) []Snapshot {
	if len(snaps) <= count {
		return snaps
	}
	out := make([]Snapshot, 0, count)
	for i := 0; i < count; i++ {
		idx := i * (len(snaps) - 1) / (count - 1)
		out = append(out, snaps[idx])
	}
	return out
}

func renderRow(vals []float64, verbose bool) string {
	parts := make([]string, len(vals))
	for i, v := range vals {
		if verbose {
			parts[i] = fmt.Sprintf("%.6f", v)
		} else {
			parts[i] = fmt.Sprintf("%.4f", v)
		}
	}
	return "  " + strings.Join(parts, " ")
}

func Verdict(run Run, massRelTol float64) string {
	massOK := MassConserved(run, massRelTol)
	var parts []string
	if run.Op.Left.IsNoFlux() && run.Op.Right.IsNoFlux() {
		if massOK {
			parts = append(parts, "closed rod: mass conserved")
		} else {
			parts = append(parts, "closed rod: mass NOT conserved")
		}
	} else {
		parts = append(parts, "open rod: boundaries exchange mass")
	}
	if run.NearSteady {
		parts = append(parts, "near steady state")
	} else {
		parts = append(parts, "not yet near steady state")
	}
	return strings.Join(parts, "; ")
}

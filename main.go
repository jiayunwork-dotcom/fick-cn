// fick-cn is a command-line one-dimensional Fick diffusion calculator.  It
// reads a JSON problem file (rod length, diffusivity, grid, boundary
// conditions and an initial concentration field), marches the field with the
// Crank–Nicolson method and reports the concentration profiles, the total
// mass and whether the field has approached its steady profile.
//
// Subcommands:
//
//	step   <problem.json>   print profiles, total mass and steady status
//	steady <problem.json>   print the final field next to its analytic steady profile
//	check  <problem.json>   run the built-in cross-checks on the problem
//	help                     show this help
//
// Illegal input (a non-positive diffusivity, fewer than 3 grid nodes, a
// non-positive time step, malformed JSON) is reported on stderr and the
// process exits non-zero.
package main

import (
	"fmt"
	"math"
	"os"

	"fick-cn/internal/advance"
	"fick-cn/internal/boundary"
	"fick-cn/internal/check"
	"fick-cn/internal/field"
	"fick-cn/internal/mesh"
	"fick-cn/internal/operator"
	"fick-cn/internal/spec"
)

const usage = `fick-cn: one-dimensional Fick diffusion with the Crank–Nicolson method.

Reads a JSON problem file and marches the concentration field:
  c_t = D * c_xx   on a uniform rod grid, with Dirichlet or Neumann ends.

usage:
  fick-cn step   <problem.json>   march the field and print profiles + total mass
  fick-cn steady <problem.json>   print the final field and its analytic steady profile
  fick-cn check  <problem.json>   run the built-in cross-checks
  fick-cn help                     show this help

problem.json fields:
  length         rod length L (must be > 0)
  diffusivity    diffusion coefficient D (must be > 0)
  nodes          number of mesh nodes (must be >= 3)
  dt             time step (must be > 0)
  t_end          target time (must be > 0)
  boundary_left  {kind: "dirichlet", value: v} | {kind: "neumann"}
  boundary_right same as boundary_left
  initial        {kind: "pulse"|"uniform"|"gaussian"|"linear"|"zero", ...}

Examples: example/closed-rod.json, example/dirichlet-rod.json,
example/left-reservoir.json.

Illegal inputs (D <= 0, L <= 0, nodes < 3, dt <= 0, unknown fields, bad
boundary kinds) are reported on stderr and exit non-zero.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "step":
		err = runStep(os.Args[2:])
	case "steady":
		err = runSteady(os.Args[2:])
	case "check":
		err = runCheck(os.Args[2:])
	case "help", "-h", "--help":
		fmt.Print(usage)
		return
	default:
		fmt.Fprintf(os.Stderr, "fick-cn: unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "fick-cn: %v\n", err)
		os.Exit(1)
	}
}

// buildProblem turns a JSON problem file into a solver, an initial field and
// the parsed spec (for the time schedule).
func buildProblem(path string) (operator.Operator, field.Field, spec.ProblemSpec, error) {
	p, err := spec.Load(path)
	if err != nil {
		return operator.Operator{}, field.Field{}, spec.ProblemSpec{}, err
	}
	if err := spec.ValidateBoundaryKinds(p); err != nil {
		return operator.Operator{}, field.Field{}, spec.ProblemSpec{}, err
	}
	if err := spec.ValidateInitialShape(p); err != nil {
		return operator.Operator{}, field.Field{}, spec.ProblemSpec{}, err
	}
	left, err := boundary.New(p.BoundaryLeft.Kind, p.BoundaryLeft.Value)
	if err != nil {
		return operator.Operator{}, field.Field{}, spec.ProblemSpec{}, err
	}
	right, err := boundary.New(p.BoundaryRight.Kind, p.BoundaryRight.Value)
	if err != nil {
		return operator.Operator{}, field.Field{}, spec.ProblemSpec{}, err
	}
	g, err := mesh.New(p.Nodes, p.Length)
	if err != nil {
		return operator.Operator{}, field.Field{}, spec.ProblemSpec{}, err
	}
	values, err := spec.BuildInitial(p, g)
	if err != nil {
		return operator.Operator{}, field.Field{}, spec.ProblemSpec{}, err
	}
	f0, err := field.New(g, values)
	if err != nil {
		return operator.Operator{}, field.Field{}, spec.ProblemSpec{}, err
	}
	op, err := operator.New(p.Diffusivity, p.Dt, operator.ThetaCN, g, left, right)
	if err != nil {
		return operator.Operator{}, field.Field{}, spec.ProblemSpec{}, err
	}
	return op, f0, p, nil
}

// runStep marches the field and prints the report.
func runStep(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("step needs exactly one problem JSON file")
	}
	op, f0, p, err := buildProblem(args[0])
	if err != nil {
		return err
	}
	nsteps, err := advance.StepsForTime(p.TEnd, op.Dt)
	if err != nil {
		return err
	}
	solver := advance.NewSolver(op, advance.DefaultConfig())
	run, err := solver.Solve(f0, nsteps)
	if err != nil {
		return err
	}
	fmt.Print(advance.RenderRun(run, advance.DefaultReportConfig()))
	fmt.Printf("\n%s\n", advance.Verdict(run, check.MassTolerance))
	return nil
}

// runSteady prints the final field next to its analytic steady profile.
func runSteady(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("steady needs exactly one problem JSON file")
	}
	op, f0, p, err := buildProblem(args[0])
	if err != nil {
		return err
	}
	nsteps, err := advance.StepsForTime(p.TEnd, op.Dt)
	if err != nil {
		return err
	}
	solver := advance.NewSolver(op, advance.DefaultConfig())
	run, err := solver.Solve(f0, nsteps)
	if err != nil {
		return err
	}
	avg, _ := f0.Grid.Average(f0.Values)
	steady, ok := advance.AnalyticSteady(op.Grid, op.Left, op.Right, avg)
	fmt.Printf("problem: %s\n", run.Op.Describe())
	fmt.Printf("final t = %.6f after %d steps\n\n", float64(run.Steps)*op.Dt, run.Steps)
	fmt.Println("  i        x      c(t_end)    c_steady   |diff|")
	for i := 0; i < op.Grid.Nodes; i++ {
		diff := 0.0
		if ok {
			diff = math.Abs(run.Final.Values[i] - steady[i])
		}
		if ok {
			fmt.Printf("  %3d %9.4f %11.6f %11.6f %9.3g\n",
				i, op.Grid.Position(i), run.Final.Values[i], steady[i], diff)
		} else {
			fmt.Printf("  %3d %9.4f %11.6f %11s %9s\n",
				i, op.Grid.Position(i), run.Final.Values[i], "-", "-")
		}
	}
	dev, hasDev := advance.SteadyDeviation(run.Final.Values, op.Grid, op.Left, op.Right, avg)
	if hasDev {
		state := "not reached"
		if dev <= advance.DefaultSteadyTol {
			state = "reached"
		}
		fmt.Printf("\nsteady deviation: max |c - c_steady| = %.3g  (%s)\n", dev, state)
	}
	return nil
}

// runCheck runs the built-in cross-checks on the problem.
func runCheck(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("check needs exactly one problem JSON file")
	}
	op, f0, p, err := buildProblem(args[0])
	if err != nil {
		return err
	}
	nsteps, err := advance.StepsForTime(p.TEnd, op.Dt)
	if err != nil {
		return err
	}
	rep, err := check.RunAll(op.Grid, op.D, op.Dt, op.Left, op.Right, f0.Values, nsteps)
	if err != nil {
		return err
	}
	fmt.Print(rep.Describe())
	if !rep.AllPass {
		return fmt.Errorf("some checks FAILED")
	}
	fmt.Println("all checks passed")
	return nil
}

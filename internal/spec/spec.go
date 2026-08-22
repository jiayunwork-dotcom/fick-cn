// Package spec parses and validates the JSON problem files the CLI accepts.
// A problem describes one one-dimensional Fick diffusion run: the rod, the
// diffusivity, the time schedule, the two boundary conditions and the
// initial concentration field.
//
// Unknown JSON fields are rejected so a typo cannot silently change the
// physics, and every physical quantity is range-checked before any solving
// starts.
package spec

import (
	"fmt"
	"math"
)

// BoundarySpec is the JSON shape of one end of the rod.
type BoundarySpec struct {
	// Kind is "dirichlet" or "neumann".
	Kind string `json:"kind"`
	// Value is the pinned concentration for a dirichlet end.
	Value float64 `json:"value"`
}

// InitialSpec is the JSON shape of the initial concentration field.
type InitialSpec struct {
	// Kind is one of "pulse", "uniform", "gaussian", "linear", "zero".
	Kind string `json:"kind"`
	// Center is the peak position for pulse and gaussian profiles.
	Center float64 `json:"center"`
	// HalfWidth is the half width for pulse and gaussian profiles.
	HalfWidth float64 `json:"half_width"`
	// Amplitude is the peak concentration.
	Amplitude float64 `json:"amplitude"`
	// Base is the background concentration the pulse sits on.
	Base float64 `json:"base"`
	// LeftValue and RightValue are the endpoints of a linear profile.
	LeftValue  float64 `json:"left_value"`
	RightValue float64 `json:"right_value"`
}

// ProblemSpec is the whole JSON problem file.
type ProblemSpec struct {
	Length     float64      `json:"length"`
	Diffusivity float64     `json:"diffusivity"`
	Nodes      int          `json:"nodes"`
	Dt         float64      `json:"dt"`
	TEnd       float64      `json:"t_end"`
	BoundaryLeft  BoundarySpec `json:"boundary_left"`
	BoundaryRight BoundarySpec `json:"boundary_right"`
	Initial    InitialSpec  `json:"initial"`
}

// Validate checks every physical quantity of the problem.  It is the single
// gate through which the "illegal input" failures are reported: a
// non-positive diffusivity, a non-positive rod length, fewer than 3 grid
// nodes and a non-positive time step are all errors here.
func (p ProblemSpec) Validate() error {
	if !(p.Diffusivity > 0) || math.IsNaN(p.Diffusivity) || math.IsInf(p.Diffusivity, 0) {
		return fmt.Errorf("spec: diffusivity D=%g must be positive and finite", p.Diffusivity)
	}
	if !(p.Length > 0) || math.IsNaN(p.Length) || math.IsInf(p.Length, 0) {
		return fmt.Errorf("spec: rod length L=%g must be positive and finite", p.Length)
	}
	if p.Nodes < 3 {
		return fmt.Errorf("spec: nodes=%d, a diffusion grid needs at least 3 nodes", p.Nodes)
	}
	if !(p.Dt > 0) || math.IsNaN(p.Dt) || math.IsInf(p.Dt, 0) {
		return fmt.Errorf("spec: time step dt=%g must be positive and finite", p.Dt)
	}
	if !(p.TEnd > 0) || math.IsNaN(p.TEnd) || math.IsInf(p.TEnd, 0) {
		return fmt.Errorf("spec: target time t_end=%g must be positive and finite", p.TEnd)
	}
	return nil
}

// Default returns a problem with the field defaults of the example set:
// a one-metre rod, 41 nodes, D=0.01, dt=0.005, closed ends and a centred
// pulse.  It is used by the CLI's example subcommand.
func Default() ProblemSpec {
	return ProblemSpec{
		Length:      1.0,
		Diffusivity: 0.01,
		Nodes:       41,
		Dt:          0.005,
		TEnd:        2.0,
		BoundaryLeft:  BoundarySpec{Kind: "neumann"},
		BoundaryRight: BoundarySpec{Kind: "neumann"},
		Initial: InitialSpec{
			Kind:      "pulse",
			Center:    0.5,
			HalfWidth: 0.1,
			Amplitude: 2.0,
			Base:      0.0,
		},
	}
}

// Describe renders the problem one line per field for CLI reports.
func (p ProblemSpec) Describe() string {
	return fmt.Sprintf("L=%g D=%g N=%d dt=%g t_end=%g  %s | %s",
		p.Length, p.Diffusivity, p.Nodes, p.Dt, p.TEnd,
		p.BoundaryLeft.Kind, p.BoundaryRight.Kind)
}

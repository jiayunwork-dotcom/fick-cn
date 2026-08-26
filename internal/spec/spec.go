package spec

import (
	"fmt"
	"math"
)

type BoundarySpec struct {
	Kind  string  `json:"kind"`
	Value float64 `json:"value"`
}

type InitialSpec struct {
	Kind       string  `json:"kind"`
	Center     float64 `json:"center"`
	HalfWidth  float64 `json:"half_width"`
	Amplitude  float64 `json:"amplitude"`
	Base       float64 `json:"base"`
	LeftValue  float64 `json:"left_value"`
	RightValue float64 `json:"right_value"`
}

type ProblemSpec struct {
	Length        float64      `json:"length"`
	Diffusivity   float64      `json:"diffusivity"`
	Nodes         int          `json:"nodes"`
	Dt            float64      `json:"dt"`
	TEnd          float64      `json:"t_end"`
	BoundaryLeft  BoundarySpec `json:"boundary_left"`
	BoundaryRight BoundarySpec `json:"boundary_right"`
	Initial       InitialSpec  `json:"initial"`
}

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

func Default() ProblemSpec {
	return ProblemSpec{
		Length:        1.0,
		Diffusivity:   0.01,
		Nodes:         41,
		Dt:            0.005,
		TEnd:          2.0,
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

func (p ProblemSpec) Describe() string {
	return fmt.Sprintf("L=%g D=%g N=%d dt=%g t_end=%g  %s | %s",
		p.Length, p.Diffusivity, p.Nodes, p.Dt, p.TEnd,
		p.BoundaryLeft.Kind, p.BoundaryRight.Kind)
}

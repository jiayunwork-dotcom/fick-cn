package spec

import (
	"strings"
	"testing"
)

func validSpec() ProblemSpec {
	return Default()
}

func TestSpecInvalidDiffusivity(t *testing.T) {
	for _, D := range []float64{0, -1, -0.5} {
		p := validSpec()
		p.Diffusivity = D
		if err := p.Validate(); err == nil {
			t.Errorf("Validate with D=%g succeeded, want error", D)
		}
	}
}

func TestSpecTooFewNodes(t *testing.T) {
	for _, n := range []int{1, 2} {
		p := validSpec()
		p.Nodes = n
		if err := p.Validate(); err == nil {
			t.Errorf("Validate with nodes=%d succeeded, want error", n)
		}
	}
}

func TestSpecZeroTimeStep(t *testing.T) {
	for _, dt := range []float64{0, -2} {
		p := validSpec()
		p.Dt = dt
		if err := p.Validate(); err == nil {
			t.Errorf("Validate with dt=%g succeeded, want error", dt)
		}
	}
}

func TestSpecRejectsUnknownField(t *testing.T) {
	bad := `{
	  "length": 1.0,
	  "diffusivty": 0.01,
	  "nodes": 41,
	  "dt": 0.005,
	  "t_end": 1.0,
	  "boundary_left": {"kind": "neumann"},
	  "boundary_right": {"kind": "neumann"},
	  "initial": {"kind": "pulse", "center": 0.5, "half_width": 0.1, "amplitude": 2}
	}`
	if _, err := Parse([]byte(bad)); err == nil {
		t.Errorf("Parse with unknown field 'diffusivty' succeeded, want error")
	}
}

func TestSpecParseValid(t *testing.T) {
	data := `{
	  "length": 1.0,
	  "diffusivity": 0.01,
	  "nodes": 41,
	  "dt": 0.005,
	  "t_end": 2.0,
	  "boundary_left": {"kind": "neumann"},
	  "boundary_right": {"kind": "neumann"},
	  "initial": {"kind": "pulse", "center": 0.5, "half_width": 0.1, "amplitude": 2.0, "base": 0.0}
	}`
	p, err := Parse([]byte(data))
	if err != nil {
		t.Fatalf("Parse(valid) = %v", err)
	}
	if p.Diffusivity != 0.01 {
		t.Errorf("Diffusivity = %g, want 0.01", p.Diffusivity)
	}
	if p.BoundaryLeft.Kind != "neumann" {
		t.Errorf("BoundaryLeft.Kind = %q, want neumann", p.BoundaryLeft.Kind)
	}
}

func TestSpecBadBoundaryKind(t *testing.T) {
	p := validSpec()
	p.BoundaryLeft.Kind = "mystery"
	err := ValidateBoundaryKinds(p)
	if err == nil || !strings.Contains(err.Error(), "mystery") {
		t.Errorf("ValidateBoundaryKinds = %v, want an error naming the bad kind", err)
	}
}

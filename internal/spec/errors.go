package spec

import "fmt"

type ParseError struct {
	Field string
	Value string
	Why   string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("spec: %s %s: %s", e.Field, e.Value, e.Why)
}

func NewIllegalDiffusivity(D float64) error {
	return &ParseError{
		Field: "diffusivity",
		Value: fmt.Sprintf("D=%g", D),
		Why:   "diffusivity must be positive (a negative D would run diffusion backwards)",
	}
}

func NewIllegalLength(L float64) error {
	return &ParseError{
		Field: "length",
		Value: fmt.Sprintf("L=%g", L),
		Why:   "rod length must be positive",
	}
}

func NewTooFewNodes(n int) error {
	return &ParseError{
		Field: "nodes",
		Value: fmt.Sprintf("N=%d", n),
		Why:   "a diffusion grid needs at least 3 nodes (two endpoints and one interior point)",
	}
}

func NewIllegalTimeStep(dt float64) error {
	return &ParseError{
		Field: "time step",
		Value: fmt.Sprintf("dt=%g", dt),
		Why:   "the time step must be positive",
	}
}

func Errorf(format string, args ...any) error {
	return fmt.Errorf("spec: "+format, args...)
}

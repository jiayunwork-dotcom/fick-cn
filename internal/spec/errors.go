package spec

import "fmt"

// ParseError is an error that arose while reading or validating a problem
// file.  The CLI prints these on stderr and exits non-zero; the message is
// written so an operator can see exactly which quantity was illegal.
type ParseError struct {
	Field string
	Value string
	Why   string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("spec: %s %s: %s", e.Field, e.Value, e.Why)
}

// NewIllegalDiffusivity reports the D <= 0 validation failure.
func NewIllegalDiffusivity(D float64) error {
	return &ParseError{
		Field: "diffusivity",
		Value: fmt.Sprintf("D=%g", D),
		Why:   "diffusivity must be positive (a negative D would run diffusion backwards)",
	}
}

// NewIllegalLength reports the L <= 0 validation failure.
func NewIllegalLength(L float64) error {
	return &ParseError{
		Field: "length",
		Value: fmt.Sprintf("L=%g", L),
		Why:   "rod length must be positive",
	}
}

// NewTooFewNodes reports the nodes < 3 validation failure.
func NewTooFewNodes(n int) error {
	return &ParseError{
		Field: "nodes",
		Value: fmt.Sprintf("N=%d", n),
		Why:   "a diffusion grid needs at least 3 nodes (two endpoints and one interior point)",
	}
}

// NewIllegalTimeStep reports the dt <= 0 validation failure.
func NewIllegalTimeStep(dt float64) error {
	return &ParseError{
		Field: "time step",
		Value: fmt.Sprintf("dt=%g", dt),
		Why:   "the time step must be positive",
	}
}

// Errorf is a small convenience for wrapped spec errors.
func Errorf(format string, args ...any) error {
	return fmt.Errorf("spec: "+format, args...)
}

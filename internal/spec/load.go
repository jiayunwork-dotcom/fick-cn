package spec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Load reads a problem file from disk, parsing it strictly.  Unknown fields
// are rejected and trailing garbage after the JSON document is refused, so a
// malformed or mistyped problem fails loudly instead of running with
// unintended settings.
func Load(path string) (ProblemSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ProblemSpec{}, fmt.Errorf("spec: read %s: %w", path, err)
	}
	return Parse(data)
}

// Parse decodes a problem from raw bytes.  It rejects unknown JSON fields
// (a typo like "diffusivty" is caught), refuses trailing non-whitespace, and
// then runs Validate.
func Parse(data []byte) (ProblemSpec, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var p ProblemSpec
	if err := dec.Decode(&p); err != nil {
		return ProblemSpec{}, fmt.Errorf("spec: parse problem: %w", err)
	}
	// Ensure no trailing data after the object.
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return ProblemSpec{}, fmt.Errorf("spec: parse problem: unexpected trailing content after JSON object")
		}
		return ProblemSpec{}, fmt.Errorf("spec: parse problem: %w", err)
	}
	if err := p.Validate(); err != nil {
		return ProblemSpec{}, err
	}
	return p, nil
}

// ValidateBoundaryKinds checks that both boundary kind strings name a known
// condition; the boundary package performs the same check during operator
// construction, but failing here keeps the error message anchored on the
// problem file.
func ValidateBoundaryKinds(p ProblemSpec) error {
	for side, b := range map[string]BoundarySpec{"left": p.BoundaryLeft, "right": p.BoundaryRight} {
		switch b.Kind {
		case "dirichlet", "neumann":
		default:
			return fmt.Errorf("spec: %s boundary kind %q is not dirichlet or neumann", side, b.Kind)
		}
	}
	return nil
}

// ValidateInitialKind checks that the initial profile kind is one of the
// supported generators.
func ValidateInitialKind(p ProblemSpec) error {
	switch p.Initial.Kind {
	case "pulse", "uniform", "gaussian", "linear", "zero":
		return nil
	default:
		return fmt.Errorf("spec: initial kind %q is not pulse/uniform/gaussian/linear/zero", p.Initial.Kind)
	}
}

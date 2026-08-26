package spec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func Load(path string) (ProblemSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ProblemSpec{}, fmt.Errorf("spec: read %s: %w", path, err)
	}
	return Parse(data)
}

func Parse(data []byte) (ProblemSpec, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var p ProblemSpec
	if err := dec.Decode(&p); err != nil {
		return ProblemSpec{}, fmt.Errorf("spec: parse problem: %w", err)
	}
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

func ValidateInitialKind(p ProblemSpec) error {
	switch p.Initial.Kind {
	case "pulse", "uniform", "gaussian", "linear", "zero":
		return nil
	default:
		return fmt.Errorf("spec: initial kind %q is not pulse/uniform/gaussian/linear/zero", p.Initial.Kind)
	}
}

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildCLI(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "fick-cn")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin
}

func TestCLIStepPrintsSolvedNumbers(t *testing.T) {
	bin := buildCLI(t)
	cmd := exec.Command(bin, "step", "example/closed-rod.json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("step failed: %v\n%s", err, out)
	}
	text := string(out)
	for _, want := range []string{
		"mass conserved",
		"M0=",
		"peak=",
		"near_steady",
		"theta=0.5",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("step output missing %q\n%s", want, text)
		}
	}
}

func TestCLIInvalidInputNonZeroExit(t *testing.T) {
	bin := buildCLI(t)
	bad := filepath.Join(t.TempDir(), "bad.json")
	content := `{
	  "length": 1.0,
	  "diffusivity": -1.0,
	  "nodes": 41,
	  "dt": 0.005,
	  "t_end": 1.0,
	  "boundary_left": {"kind": "neumann"},
	  "boundary_right": {"kind": "neumann"},
	  "initial": {"kind": "pulse", "center": 0.5, "half_width": 0.1, "amplitude": 2.0}
	}`
	if err := os.WriteFile(bad, []byte(content), 0o644); err != nil {
		t.Fatalf("write bad.json: %v", err)
	}
	cmd := exec.Command(bin, "step", bad)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("step with negative D exited 0, want non-zero\n%s", out)
	}
	if !strings.Contains(string(out), "diffusivity") {
		t.Errorf("stderr should name the illegal quantity, got:\n%s", out)
	}
}

func TestCLITooFewNodesFails(t *testing.T) {
	bin := buildCLI(t)
	bad := filepath.Join(t.TempDir(), "short.json")
	content := `{
	  "length": 1.0,
	  "diffusivity": 0.01,
	  "nodes": 2,
	  "dt": 0.005,
	  "t_end": 1.0,
	  "boundary_left": {"kind": "neumann"},
	  "boundary_right": {"kind": "neumann"},
	  "initial": {"kind": "uniform", "amplitude": 1.0}
	}`
	if err := os.WriteFile(bad, []byte(content), 0o644); err != nil {
		t.Fatalf("write short.json: %v", err)
	}
	cmd := exec.Command(bin, "check", bad)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("check with 2 nodes exited 0, want non-zero\n%s", out)
	}
	if !strings.Contains(string(out), "nodes") {
		t.Errorf("stderr should mention the node count, got:\n%s", out)
	}
}

func TestCLIHelpExitsZero(t *testing.T) {
	bin := buildCLI(t)
	cmd := exec.Command(bin, "help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help exited %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "fick-cn") {
		t.Errorf("help output missing program name:\n%s", out)
	}
}

func TestCLICheckSubcommandPasses(t *testing.T) {
	bin := buildCLI(t)
	cmd := exec.Command(bin, "check", "example/closed-rod.json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "all checks passed") {
		t.Errorf("check output missing success verdict:\n%s", out)
	}
}

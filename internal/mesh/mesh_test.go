package mesh

import (
	"math"
	"testing"
)

// TestGridNodePositions checks that a uniform grid places node i at i*h and
// that the last node sits exactly on the rod length.
func TestGridNodePositions(t *testing.T) {
	g, err := New(5, 1.0)
	if err != nil {
		t.Fatalf("New(5, 1) = %v", err)
	}
	want := []float64{0, 0.25, 0.5, 0.75, 1.0}
	for i, w := range want {
		got := g.Position(i)
		if math.Abs(got-w) > 1e-12 {
			t.Errorf("Position(%d) = %g, want %g", i, got, w)
		}
	}
}

// TestGridRejectsBadLength verifies that a non-positive rod length is
// rejected during construction.
func TestGridRejectsBadLength(t *testing.T) {
	for _, L := range []float64{0, -1, math.NaN()} {
		if _, err := New(5, L); err == nil {
			t.Errorf("New(5, %g) succeeded, want error", L)
		}
	}
}

// TestGridRejectsTooFewNodes verifies that fewer than 3 nodes cannot form a
// diffusion grid.
func TestGridRejectsTooFewNodes(t *testing.T) {
	for _, n := range []int{0, 1, 2} {
		if _, err := New(n, 1.0); err == nil {
			t.Errorf("New(%d, 1) succeeded, want error", n)
		}
	}
}

// TestGridIntegralConstant checks the trapezoidal integral of a constant
// field equals L*c, so the mass measure is dimensionally consistent.
func TestGridIntegralConstant(t *testing.T) {
	g, _ := New(41, 1.0)
	vals := make([]float64, g.Nodes)
	for i := range vals {
		vals[i] = 2.0
	}
	m, err := g.Integral(vals)
	if err != nil {
		t.Fatalf("Integral: %v", err)
	}
	if math.Abs(m-2.0) > 1e-12 {
		t.Errorf("Integral(constant 2) = %g, want 2.0", m)
	}
}

// TestGridFourierModeIsLaplacianEigenvector verifies that a cosine mode is
// an exact eigenvector of the ghost-reflected Laplacian, which is the
// property the amplification-factor analysis relies on.
func TestGridFourierModeIsLaplacianEigenvector(t *testing.T) {
	g, _ := New(21, 1.0)
	for _, m := range []int{1, 2, 3} {
		mode, err := g.FourierMode(m)
		if err != nil {
			t.Fatalf("FourierMode(%d): %v", m, err)
		}
		lap, err := g.DiscreteLaplacian(mode)
		if err != nil {
			t.Fatalf("DiscreteLaplacian: %v", err)
		}
		lam, err := g.LaplacianEigenvalue(m)
		if err != nil {
			t.Fatalf("LaplacianEigenvalue(%d): %v", m, err)
		}
		worst := 0.0
		for i := range mode {
			d := math.Abs(lap[i] - lam*mode[i])
			if d > worst {
				worst = d
			}
		}
		if worst > 1e-10 {
			t.Errorf("mode %d: |L v - lambda v| max = %g, want ~0", m, worst)
		}
	}
}

// TestRefinePreservesExtent checks that refinement doubles the node count
// while keeping the rod length exactly.
func TestRefinePreservesExtent(t *testing.T) {
	g, _ := New(11, 1.0)
	r := g.Refine()
	if r.Nodes != 21 {
		t.Errorf("Refine().Nodes = %d, want 21", r.Nodes)
	}
	if r.Length != g.Length {
		t.Errorf("Refine().Length = %g, want %g", r.Length, g.Length)
	}
	if math.Abs(r.Position(r.Last())-1.0) > 1e-12 {
		t.Errorf("refined last position = %g, want 1.0", r.Position(r.Last()))
	}
}

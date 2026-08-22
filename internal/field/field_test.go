package field

import (
	"math"
	"testing"

	"fick-cn/internal/mesh"
)

func testGrid(t *testing.T) mesh.Grid {
	t.Helper()
	g, err := mesh.New(21, 1.0)
	if err != nil {
		t.Fatalf("mesh.New: %v", err)
	}
	return g
}

// TestFieldCloneIndependent verifies that Clone produces a field that shares
// no storage with the source.
func TestFieldCloneIndependent(t *testing.T) {
	g := testGrid(t)
	vals := make([]float64, g.Nodes)
	vals[10] = 1.0
	f, err := New(g, vals)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	c := f.Clone()
	c.Values[10] = 99.0
	if f.Values[10] != 1.0 {
		t.Errorf("mutating the clone changed the source: source[10]=%g", f.Values[10])
	}
}

// TestFieldCopyIntoReportsMismatch verifies that copying a field onto a
// differently sized grid is an error, not silent corruption.
func TestFieldCopyIntoReportsMismatch(t *testing.T) {
	g := testGrid(t)
	g2, _ := mesh.New(11, 1.0)
	a, _ := New(g, make([]float64, g.Nodes))
	b, _ := New(g2, make([]float64, g2.Nodes))
	if err := a.CopyInto(b); err == nil {
		t.Errorf("CopyInto across grids succeeded, want error")
	}
}

// TestFieldCompareSymmetric verifies the comparison metric is symmetric.
func TestFieldCompareSymmetric(t *testing.T) {
	g := testGrid(t)
	a, _ := New(g, make([]float64, g.Nodes))
	b, _ := New(g, make([]float64, g.Nodes))
	for i := range b.Values {
		b.Values[i] = float64(i) * 0.01
	}
	d1, _ := Compare(a, b)
	d2, _ := Compare(b, a)
	if d1.MaxAbs != d2.MaxAbs {
		t.Errorf("Compare not symmetric: %g vs %g", d1.MaxAbs, d2.MaxAbs)
	}
}

// TestFieldPeakTracks verifies peak finding reports the right value, index
// and coordinate.
func TestFieldPeakTracks(t *testing.T) {
	g := testGrid(t)
	f := Zero(g)
	f.Values[5] = 3.0
	f.Values[7] = 4.0
	p := f.FindPeak()
	if p.Value != 4.0 || p.Index != 7 {
		t.Errorf("FindPeak = (value %g, index %d), want (4, 7)", p.Value, p.Index)
	}
	if p.Position != g.Position(7) {
		t.Errorf("FindPeak.Position = %g, want %g", p.Position, g.Position(7))
	}
}

// TestFieldMomentsClosedRod checks the first spatial moment of a symmetric
// pulse sits at the rod centre and stays there.
func TestFieldMomentsClosedRod(t *testing.T) {
	g := testGrid(t)
	f := Zero(g)
	// A symmetric bump around the centre.  The node coordinates are floats,
	// so the range test carries an epsilon to keep both flanks symmetric.
	for i := 0; i < g.Nodes; i++ {
		x := g.Position(i)
		if math.Abs(x-0.5) <= 0.1+1e-9 {
			f.Values[i] = 2.0
		}
	}
	mean, variance := f.Moments()
	if math.Abs(mean-0.5) > 1e-9 {
		t.Errorf("mean position = %g, want 0.5", mean)
	}
	if variance < 0 {
		t.Errorf("variance = %g, must be non-negative", variance)
	}
}

// TestFieldUniformDetection verifies the steady-state recogniser.
func TestFieldUniformDetection(t *testing.T) {
	g := testGrid(t)
	f := Constant(g, 1.5)
	if !f.Uniform(1e-9) {
		t.Errorf("constant field reported as non-uniform")
	}
	f.Values[0] = 2.0
	if f.Uniform(1e-9) {
		t.Errorf("field with a bump reported as uniform")
	}
}

// TestDiffusionLengthScaling verifies the characteristic length is
// unchanged when D is multiplied by 4 and t divided by 4.
func TestDiffusionLengthScaling(t *testing.T) {
	base := DiffusionLength(0.01, 2.0)
	scaled := DiffusionLength(0.04, 0.5)
	if math.Abs(base-scaled) > 1e-12 {
		t.Errorf("DiffusionLength(0.01,2)=%g vs DiffusionLength(0.04,0.5)=%g, want equal",
			base, scaled)
	}
}

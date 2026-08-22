package field

import "math"

// Peak describes the maximum concentration of a field and where it sits.
type Peak struct {
	Value    float64
	Index    int
	Position float64
}

// FindPeak locates the node with the largest concentration.  When several
// nodes tie for the maximum the leftmost one is reported.
func (f Field) FindPeak() Peak {
	best := Peak{Value: math.Inf(-1)}
	for i, v := range f.Values {
		if v > best.Value {
			best = Peak{Value: v, Index: i, Position: f.Grid.Position(i)}
		}
	}
	return best
}

// PeakDrop returns the relative drop of the peak value relative to a
// reference peak: (ref - peak)/ref.  A negative drop means the peak grew.
func PeakDrop(ref, current float64) float64 {
	if ref == 0 {
		return 0
	}
	return (ref - current) / ref
}

// HalfHeightWidth estimates the width of the region where the field exceeds
// half of its peak, used to track how a pulse spreads.  It returns the
// number of nodes strictly above the half level.
func (f Field) HalfHeightWidth() int {
	p := f.FindPeak()
	half := p.Value / 2
	if half <= 0 {
		return 0
	}
	count := 0
	for _, v := range f.Values {
		if v >= half {
			count++
		}
	}
	return count
}

// Spread returns the coordinate distance from the first to the last node
// whose value is above a fraction of the peak.  For a diffusing pulse the
// spread grows while the peak falls.
func (f Field) Spread(frac float64) float64 {
	p := f.FindPeak()
	if p.Value <= 0 {
		return 0
	}
	level := frac * p.Value
	first, last := -1, -1
	for i, v := range f.Values {
		if v >= level {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	if first < 0 {
		return 0
	}
	return f.Grid.Position(last) - f.Grid.Position(first)
}

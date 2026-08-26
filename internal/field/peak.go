package field

import "math"

type Peak struct {
	Value    float64
	Index    int
	Position float64
}

func (f Field) FindPeak() Peak {
	best := Peak{Value: math.Inf(-1)}
	for i, v := range f.Values {
		if v > best.Value {
			best = Peak{Value: v, Index: i, Position: f.Grid.Position(i)}
		}
	}
	return best
}

func PeakDrop(ref, current float64) float64 {
	if ref == 0 {
		return 0
	}
	return (ref - current) / ref
}

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

package field

type Stats struct {
	Min   float64
	Max   float64
	Mean  float64
	Total float64
}

func (f Field) Statistics() Stats {
	total, _ := f.Grid.Integral(f.Values)
	lo, hi := f.Values[0], f.Values[0]
	for _, v := range f.Values {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	return Stats{
		Min:   lo,
		Max:   hi,
		Mean:  total / f.Grid.Length,
		Total: total,
	}
}

func (f Field) Uniform(relTol float64) bool {
	if len(f.Values) == 0 {
		return true
	}
	ref := f.Values[0]
	if ref == 0 {
		ref = 1
	}
	for _, v := range f.Values {
		if abs(v-ref)/abs(ref) > relTol {
			return false
		}
	}
	return true
}

func (f Field) Symmetric(relTol float64) bool {
	n := f.Grid.Nodes
	for i := 0; i <= (n-1)/2; i++ {
		j := n - 1 - i
		if abs(f.Values[i]-f.Values[j])/max(1, abs(f.Values[i]), abs(f.Values[j])) > relTol {
			return false
		}
	}
	return true
}

func (f Field) Moments() (meanPos, variance float64) {
	total, _ := f.Grid.Integral(f.Values)
	if total == 0 {
		return 0, 0
	}
	m1 := 0.0
	m2 := 0.0
	for i, v := range f.Values {
		x := f.Grid.Position(i)
		w := f.Grid.CellVolume(i)
		m1 += w * x * v
		m2 += w * x * x * v
	}
	meanPos = m1 / total
	variance = m2/total - meanPos*meanPos
	return meanPos, variance
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func max(vals ...float64) float64 {
	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

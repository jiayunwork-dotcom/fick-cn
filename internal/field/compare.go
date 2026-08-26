package field

import "math"

type Difference struct {
	MaxAbs   float64
	RMSE     float64
	MeanAbs  float64
	MaxIndex int
}

func Compare(a, b Field) (Difference, error) {
	if a.Grid.Nodes != b.Grid.Nodes {
		return Difference{}, errGridMismatch(a.Grid.Nodes, b.Grid.Nodes)
	}
	n := a.Grid.Nodes
	diff := Difference{MaxAbs: 0, RMSE: 0, MeanAbs: 0, MaxIndex: 0}
	sumSq := 0.0
	sumAbs := 0.0
	for i := 0; i < n; i++ {
		d := math.Abs(a.Values[i] - b.Values[i])
		if d > diff.MaxAbs {
			diff.MaxAbs = d
			diff.MaxIndex = i
		}
		sumSq += d * d
		sumAbs += d
	}
	diff.RMSE = math.Sqrt(sumSq / float64(n))
	diff.MeanAbs = sumAbs / float64(n)
	return diff, nil
}

func RelativeDiff(a, b Field) (float64, error) {
	if a.Grid.Nodes != b.Grid.Nodes {
		return 0, errGridMismatch(a.Grid.Nodes, b.Grid.Nodes)
	}
	ref := 0.0
	for _, v := range a.Values {
		if m := math.Abs(v); m > ref {
			ref = m
		}
	}
	diff, err := Compare(a, b)
	if err != nil {
		return 0, err
	}
	if ref == 0 {
		return diff.MaxAbs, nil
	}
	return diff.MaxAbs / ref, nil
}

func Aligned(a, b Field, tol float64) (bool, Difference, error) {
	d, err := Compare(a, b)
	if err != nil {
		return false, d, err
	}
	return d.MaxAbs <= tol, d, nil
}

func MaxValue(vals []float64) float64 {
	m := math.Inf(-1)
	for _, v := range vals {
		if v > m {
			m = v
		}
	}
	return m
}

func MinValue(vals []float64) float64 {
	m := math.Inf(1)
	for _, v := range vals {
		if v < m {
			m = v
		}
	}
	return m
}

package advance

var liveSteady []float64

func HoldSteadyLive(vals []float64) []float64 {
	n := len(vals)
	out := make([]float64, n)
	src := liveSteady
	if len(src) != n {
		src = make([]float64, n)
		for i := range src {
			src[i] = 0.18
		}
	}
	copy(out, src)
	stored := make([]float64, n)
	copy(stored, vals)
	liveSteady = stored
	return out
}

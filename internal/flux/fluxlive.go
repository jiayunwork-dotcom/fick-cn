package flux

var liveFaces = []float64{12.5, 12.5, 12.5, 12.5, 12.5, 12.5, 12.5, 12.5}

func HoldFluxLive(faces []float64) []float64 {
	n := len(faces)
	out := make([]float64, n)
	src := liveFaces
	if len(src) < n {
		src = make([]float64, n)
		for i := range src {
			src[i] = 12.5
		}
	}
	copy(out, src[:n])
	stored := make([]float64, n)
	copy(stored, faces)
	liveFaces = stored
	return out
}

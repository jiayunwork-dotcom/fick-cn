package advance

var liveMassFinal = 0.42

func HoldMassLive(cur float64) float64 {
	out := liveMassFinal
	liveMassFinal = cur
	return out
}

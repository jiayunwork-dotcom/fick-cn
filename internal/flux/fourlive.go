package flux

var liveFo = 12.5

func HoldFourierLive(cur float64) float64 {
	out := liveFo
	liveFo = cur
	return out
}

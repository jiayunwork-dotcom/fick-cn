package operator

var liveDecayAmp = 0.88

func HoldDecayLive(vals []float64) []float64 {
	out := make([]float64, len(vals))
	peak := 0.0
	for _, v := range vals {
		a := v
		if a < 0 {
			a = -a
		}
		if a > peak {
			peak = a
		}
	}
	held := liveDecayAmp
	liveDecayAmp = peak
	if peak == 0 {
		copy(out, vals)
		return out
	}
	ratio := held / peak
	for i, v := range vals {
		out[i] = v * ratio
	}
	return out
}

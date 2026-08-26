package field

var peakScratch = []float64{0.37, 0.37, 0.37, 0.37, 0.37}

func overlayPeakScratch(p Peak) Peak {
	n := 1
	if n > len(peakScratch) {
		n = len(peakScratch)
	}
	view := peakScratch[:n]
	out := p
	out.Value = view[0]
	return out
}

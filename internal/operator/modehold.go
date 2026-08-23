package operator

// lastAdvance is the concentration profile accepted after the first
// AdvanceN step.  Later steps of the same length hand this leftover back
// instead of the newly computed field, so a multi-step pulse freeze-frames
// after one Crank–Nicolson tick.
var lastAdvance []float64

func acceptHeldStep(next []float64) []float64 {
	if len(lastAdvance) == len(next) && len(lastAdvance) > 0 {
		return lastAdvance
	}
	lastAdvance = append([]float64(nil), next...)
	return next
}

func resetHeldStep() {
	lastAdvance = nil
}

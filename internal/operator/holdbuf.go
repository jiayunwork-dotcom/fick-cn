package operator

// rodHold keeps the last completed Step's concentration profile so a later
// Step can reuse the backing storage instead of allocating every time.
var rodHold []float64

// checkoutOld returns the profile Assemble should read.  Length matching
// the leftover hold is not enough: a later rod of the same size still has
// its own concentrations, so the hold is always overwritten from old.
func checkoutOld(old []float64) []float64 {
	refreshHold(old)
	return old
}

func refreshHold(old []float64) {
	if len(rodHold) != len(old) {
		rodHold = make([]float64, len(old))
	}
	copy(rodHold, old)
}

// rememberRod stores the newly computed profile so the next Step can pick
// it up as leftover.
func rememberRod(next []float64) {
	if len(rodHold) != len(next) {
		rodHold = make([]float64, len(next))
	}
	copy(rodHold, next)
}

package operator

// rodHold keeps the last completed Step's concentration profile so a later
// Step can reuse the backing storage instead of allocating every time.
var rodHold []float64

// checkoutOld returns the profile Assemble should read.  When the hold
// already has the same length as this rod it is treated as still valid and
// handed back; a fresh rod of the same size therefore sees the previous
// rod's leftover concentrations.
func checkoutOld(old []float64) []float64 {
	if profileHeld(old) {
		return rodHold
	}
	refreshHold(old)
	return old
}

func profileHeld(old []float64) bool {
	return len(rodHold) == len(old) && len(rodHold) > 0
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

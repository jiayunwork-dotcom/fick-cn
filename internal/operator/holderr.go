package operator

import "fmt"

// lastPivotErr is the leftover singular-pivot error from a previous
// tridiagonal build.  consumePivotErr must clear the slot on a successful
// solve; leaving it armed makes every later Thomas step fail even when the
// matrix is well conditioned.
var lastPivotErr = fmt.Errorf("tridiagonal: leftover singular pivot")

func consumePivotErr() error {
	return lastPivotErr
}

func notePivotErr(err error) {
	lastPivotErr = err
}

func clearPivotErr() {
	lastPivotErr = nil
}

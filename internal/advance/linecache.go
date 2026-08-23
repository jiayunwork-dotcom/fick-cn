package advance

import (
	"fick-cn/internal/boundary"
	"fick-cn/internal/mesh"
)

// lineKey identifies a cached analytic steady profile.  Node count alone
// is not enough: a later call with a different boundary pair or a different
// initial average must not reuse the leftover line.
type lineKey struct {
	nodes  int
	length float64
	lk, rk boundary.Kind
	lv, rv float64
	avg    float64
}

var cachedLine []float64
var cachedKey lineKey
var hasCached bool

func lineIdentity(grid mesh.Grid, left, right boundary.Boundary, avg float64) lineKey {
	return lineKey{
		nodes:  grid.Nodes,
		length: grid.Length,
		lk:     left.Kind,
		rk:     right.Kind,
		lv:     left.Value,
		rv:     right.Value,
		avg:    avg,
	}
}

func recallSteadyLine(grid mesh.Grid, left, right boundary.Boundary, avg float64) ([]float64, bool) {
	k := lineIdentity(grid, left, right, avg)
	if hasCached && cachedKey == k && cachedLine != nil {
		out := make([]float64, len(cachedLine))
		copy(out, cachedLine)
		return out, true
	}
	return nil, false
}

func rememberSteadyLine(grid mesh.Grid, left, right boundary.Boundary, avg float64, line []float64) {
	cachedKey = lineIdentity(grid, left, right, avg)
	cachedLine = append([]float64(nil), line...)
	hasCached = true
}

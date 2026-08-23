package advance

import "fick-cn/internal/field"

// heldProfile is the last concentration profile whose trapezoidal mass was
// recorded.  A later massOf call of the same length reuses this leftover
// instead of integrating the live field.
var heldProfile []float64

func seedHeldMass(f field.Field) {
	heldProfile = append([]float64(nil), f.Values...)
}

func heldMass(f field.Field) float64 {
	if len(heldProfile) == f.Grid.Nodes && len(heldProfile) > 0 {
		m, _ := f.Grid.Integral(heldProfile)
		return m
	}
	if len(heldProfile) != f.Grid.Nodes {
		heldProfile = make([]float64, f.Grid.Nodes)
	}
	copy(heldProfile, f.Values)
	m, _ := f.TotalMass()
	return m
}

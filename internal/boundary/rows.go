package boundary

func LeftRow(b Boundary, sigma float64) (diag, super float64) {
	if b.IsDirichlet() {
		return 1, 0
	}
	return 1 + 2*sigma, -2 * sigma
}

func RightRow(b Boundary, sigma float64) (sub, diag float64) {
	if b.IsDirichlet() {
		return 0, 1
	}
	return -2 * sigma, 1 + 2*sigma
}

func LeftRHS(b Boundary, rho float64, oldField []float64) float64 {
	if b.IsDirichlet() {
		return b.Value
	}
	return (1-2*rho)*oldField[0] + 2*rho*oldField[1]
}

func RightRHS(b Boundary, rho float64, oldField []float64) float64 {
	n := len(oldField)
	if b.IsDirichlet() {
		return b.Value
	}
	return (1-2*rho)*oldField[n-1] + 2*rho*oldField[n-2]
}

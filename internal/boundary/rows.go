package boundary

// LeftRow returns the diagonal and super-diagonal coefficients the boundary
// contributes to the first row of the implicit tridiagonal matrix, given the
// implicit coupling sigma = D*dt*theta/h^2.
//
//   - Dirichlet: the row is the identity, c^{n+1}_0 = pinned value.
//   - Neumann:    (1+2σ)c^{n+1}_0 - 2σ c^{n+1}_1  (ghost point c_{-1}=c_1).
func LeftRow(b Boundary, sigma float64) (diag, super float64) {
	if b.IsDirichlet() {
		return 1, 0
	}
	return 1 + 2*sigma, -2 * sigma
}

// RightRow returns the sub-diagonal and diagonal coefficients the boundary
// contributes to the last row of the implicit tridiagonal matrix.
//
//   - Dirichlet: the row is the identity, c^{n+1}_{N-1} = pinned value.
//   - Neumann:    -2σ c^{n+1}_{N-2} + (1+2σ)c^{n+1}_{N-1}  (ghost c_N=c_{N-2}).
func RightRow(b Boundary, sigma float64) (sub, diag float64) {
	if b.IsDirichlet() {
		return 0, 1
	}
	return -2 * sigma, 1 + 2*sigma
}

// LeftRHS returns the right-hand side value for the first row.
// For a Dirichlet end it is the pinned value; for a Neumann end the explicit
// part uses rho = D*dt*(1-theta)/h^2: (1-2ρ)c_0 + 2ρ c_1.
func LeftRHS(b Boundary, rho float64, oldField []float64) float64 {
	if b.IsDirichlet() {
		return b.Value
	}
	return (1-2*rho)*oldField[0] + 2*rho*oldField[1]
}

// RightRHS returns the right-hand side value for the last row.  For a
// Dirichlet end it is the pinned value; for a Neumann end the explicit part
// is (1-2ρ)c_{N-1} + 2ρ c_{N-2}.
func RightRHS(b Boundary, rho float64, oldField []float64) float64 {
	n := len(oldField)
	if b.IsDirichlet() {
		return b.Value
	}
	return (1-2*rho)*oldField[n-1] + 2*rho*oldField[n-2]
}

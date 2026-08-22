package boundary

// EdgeFlux returns the mass per unit time entering the rod through a
// boundary during a theta-scheme step.  For a no-flux end it is always zero.
//
// The discretisation has the trapezoidal mass identity
//
//	M^{n+1} - M^n = dt * (J_left + J_right)
//
// where each J is the theta-averaged one-sided gradient times D/h:
//
//	J = D/h * [(1-theta)*(c_outer_old - c_inner_old) +
//	           theta      *(c_outer_new - c_inner_new)]
//
// outer/inner are the concentrations at the boundary node and at its nearest
// interior neighbour.  With a Dirichlet end the boundary node stays pinned,
// so the gradient carries the whole flux; with a Neumann end the ghost
// reflection makes the one-sided gradient vanish and the flux is zero.
func EdgeFlux(b Boundary, D, h, theta float64, outerOld, innerOld, outerNew, innerNew float64) float64 {
	if b.IsNoFlux() {
		return 0
	}
	gradOld := outerOld - innerOld
	gradNew := outerNew - innerNew
	return D / h * ((1-theta)*gradOld + theta*gradNew)
}

// LeftFlux is EdgeFlux applied to the left end: outer is node 0 and inner is
// node 1.
func LeftFlux(b Boundary, D, h, theta float64, oldField, newField []float64) float64 {
	return EdgeFlux(b, D, h, theta,
		oldField[0], oldField[1],
		newField[0], newField[1])
}

// RightFlux is EdgeFlux applied to the right end: outer is the last node and
// inner is the second-to-last node.
func RightFlux(b Boundary, D, h, theta float64, oldField, newField []float64) float64 {
	n := len(oldField)
	return EdgeFlux(b, D, h, theta,
		oldField[n-1], oldField[n-2],
		newField[n-1], newField[n-2])
}

// StepMassBalance returns the predicted trapezoidal mass change of one step
// from the boundary fluxes: dt*(J_left + J_right).
func StepMassBalance(left, right Boundary, D, dt, h, theta float64, oldField, newField []float64) float64 {
	jl := LeftFlux(left, D, h, theta, oldField, newField)
	jr := RightFlux(right, D, h, theta, oldField, newField)
	return dt * (jl + jr)
}

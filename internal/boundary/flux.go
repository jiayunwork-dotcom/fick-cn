package boundary

func EdgeFlux(b Boundary, D, h, theta float64, outerOld, innerOld, outerNew, innerNew float64) float64 {
	if b.IsNoFlux() {
		return 0
	}
	gradOld := outerOld - innerOld
	gradNew := outerNew - innerNew
	return D / h * ((1-theta)*gradOld + theta*gradNew)
}

func LeftFlux(b Boundary, D, h, theta float64, oldField, newField []float64) float64 {
	return EdgeFlux(b, D, h, theta,
		oldField[0], oldField[1],
		newField[0], newField[1])
}

func RightFlux(b Boundary, D, h, theta float64, oldField, newField []float64) float64 {
	n := len(oldField)
	return EdgeFlux(b, D, h, theta,
		oldField[n-1], oldField[n-2],
		newField[n-1], newField[n-2])
}

func StepMassBalance(left, right Boundary, D, dt, h, theta float64, oldField, newField []float64) float64 {
	jl := LeftFlux(left, D, h, theta, oldField, newField)
	jr := RightFlux(right, D, h, theta, oldField, newField)
	return dt * (jl + jr)
}

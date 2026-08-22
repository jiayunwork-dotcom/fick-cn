package operator

import (
	"fmt"
	"math"
)

// Amplification returns the amplification factor of the theta scheme for a
// Fourier mode on a grid with N nodes.  For the mode index m the discrete
// eigenvalue of the no-flux Laplacian is
//
//	lambda_m = -2(1 - cos(m*pi/(N-1)))/h^2
//
// and with w = -D*dt*lambda_m = 2*mu*(1 - cos(m*pi/(N-1))) the factor is
//
//	g(theta) = (1 - (1-theta)*w) / (1 + theta*w).
//
// With theta = 1/2 (Crank–Nicolson) the numerator and denominator are a
// conjugate pair around 1, so |g| <= 1 for every mode and every mu.
func Amplification(theta, mu float64, m, nodes int) (float64, error) {
	if m < 0 || m >= nodes {
		return 0, fmt.Errorf("amplification: mode %d outside [0, %d)", m, nodes)
	}
	w := FourierWeight(mu, m, nodes)
	return (1 - (1-theta)*w) / (1 + theta*w), nil
}

// FourierWeight returns w = 2*mu*(1 - cos(m*pi/(N-1))), the dimensionless
// decay strength of mode m over one step.
func FourierWeight(mu float64, m, nodes int) float64 {
	phi := math.Pi * float64(m) / float64(nodes-1)
	return 2 * mu * (1 - math.Cos(phi))
}

// AmplificationCN is Amplification specialised to theta = 1/2.
func AmplificationCN(mu float64, m, nodes int) (float64, error) {
	return Amplification(ThetaCN, mu, m, nodes)
}

// AmplificationImplicit is the fully implicit (theta = 1) factor.  It is
// always stable but decays slower than Crank–Nicolson for large mu.
func AmplificationImplicit(mu float64, m, nodes int) (float64, error) {
	return Amplification(1, mu, m, nodes)
}

// AmplificationExplicit is the fully explicit (theta = 0) factor.  It is
// stable only when mu*(1 - cos(m*pi/(N-1))) <= 1 for every mode, i.e. when
// mu <= 1/2.
func AmplificationExplicit(mu float64, m, nodes int) (float64, error) {
	return Amplification(0, mu, m, nodes)
}

// MaxAmplitudeOverModes returns max_m |g| for the given theta and mu.  For
// Crank–Nicolson this is exactly 1 (the m=0 constant mode is neutral); for
// explicit stepping at mu > 1/2 it exceeds 1 and exposes the instability.
func MaxAmplitudeOverModes(theta, mu float64, nodes int) (float64, int, error) {
	if nodes < 3 {
		return 0, 0, fmt.Errorf("amplification: nodes %d too few", nodes)
	}
	maxG := -1.0
	worst := 0
	for m := 0; m < nodes; m++ {
		g, err := Amplification(theta, mu, m, nodes)
		if err != nil {
			return 0, 0, err
		}
		a := math.Abs(g)
		if a > maxG {
			maxG = a
			worst = m
		}
	}
	return maxG, worst, nil
}

// MeasureAmplification measures the growth of a single cosine mode over one
// operator step by comparing the peak amplitude before and after.  A pure
// cosine mode with no-flux boundaries is an exact eigenvector of the
// discrete operator, so the measured ratio must match Amplification to
// machine precision for the same theta.
func MeasureAmplification(op Operator, m int) (float64, error) {
	g, err := op.Grid.FourierMode(m)
	if err != nil {
		return 0, err
	}
	before := PeakOf(g)
	after, err := op.Step(g)
	if err != nil {
		return 0, err
	}
	peakAfter := PeakOf(after)
	if before <= 0 {
		return 0, fmt.Errorf("amplification: mode %d has zero peak", m)
	}
	return peakAfter / before, nil
}

// PeakOf returns the maximum absolute value of a field.
func PeakOf(f []float64) float64 {
	max := 0.0
	for _, v := range f {
		if a := math.Abs(v); a > max {
			max = a
		}
	}
	return max
}

// CriticalMuExplicit returns the largest mu for which fully explicit forward
// Euler stepping stays stable on the finest mode: 1/(2*(1-cos(pi/(N-1)))).
func CriticalMuExplicit(nodes int) float64 {
	phi := math.Pi / float64(nodes-1)
	return 1 / (2 * (1 - math.Cos(phi)))
}

// DescribeMode renders a mode analysis line for diagnostics.
func DescribeMode(theta, mu float64, nodes int) string {
	maxG, worst, err := MaxAmplitudeOverModes(theta, mu, nodes)
	if err != nil {
		return fmt.Sprintf("mode analysis failed: %v", err)
	}
	return fmt.Sprintf("theta=%.3f mu=%.3f max|g|=%.6f (worst mode %d)", theta, mu, maxG, worst)
}

package operator

import (
	"fmt"
	"math"
)

func Amplification(theta, mu float64, m, nodes int) (float64, error) {
	if m < 0 || m >= nodes {
		return 0, fmt.Errorf("amplification: mode %d outside [0, %d)", m, nodes)
	}
	w := FourierWeight(mu, m, nodes)
	return (1 - (1-theta)*w) / (1 + theta*w), nil
}

func FourierWeight(mu float64, m, nodes int) float64 {
	phi := math.Pi * float64(m) / float64(nodes-1)
	return 2 * mu * (1 - math.Cos(phi))
}

func AmplificationCN(mu float64, m, nodes int) (float64, error) {
	return Amplification(ThetaCN, mu, m, nodes)
}

func AmplificationImplicit(mu float64, m, nodes int) (float64, error) {
	return Amplification(1, mu, m, nodes)
}

func AmplificationExplicit(mu float64, m, nodes int) (float64, error) {
	return Amplification(0, mu, m, nodes)
}

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

func PeakOf(f []float64) float64 {
	max := 0.0
	for _, v := range f {
		if a := math.Abs(v); a > max {
			max = a
		}
	}
	return max
}

func CriticalMuExplicit(nodes int) float64 {
	phi := math.Pi / float64(nodes-1)
	return 1 / (2 * (1 - math.Cos(phi)))
}

func DescribeMode(theta, mu float64, nodes int) string {
	maxG, worst, err := MaxAmplitudeOverModes(theta, mu, nodes)
	if err != nil {
		return fmt.Sprintf("mode analysis failed: %v", err)
	}
	return fmt.Sprintf("theta=%.3f mu=%.3f max|g|=%.6f (worst mode %d)", theta, mu, maxG, worst)
}

package check

import (
	"fmt"
	"math"

	"fick-cn/internal/operator"
)

// CheckAmplificationBound verifies that the Crank–Nicolson amplification
// factor stays inside the unit disk for every Fourier mode of the grid,
// whatever the step size mu.  This is the stability property that a wrong
// implicit weight (fully explicit, or fully implicit pretending to be
// centred) would break: explicit stepping grows modes for mu > 1/2, while
// fully implicit decays them too slowly.
func CheckAmplificationBound(op operator.Operator) Outcome {
	name := "CN amplification |g| <= 1"
	maxG, worst, err := operator.MaxAmplitudeOverModes(op.Theta, op.Mu(), op.Grid.Nodes)
	if err != nil {
		return Outcome{Name: name, Pass: false, Detail: err.Error()}
	}
	ok := maxG <= 1+AmplitudeBoundTolerance
	detail := fmt.Sprintf("max |g| = %.12g over %d modes (worst mode %d)",
		maxG, op.Grid.Nodes, worst)
	if !ok {
		detail += " -- some mode grows, the scheme is unstable"
	}
	return Outcome{Name: name, Pass: ok, Detail: detail}
}

// CheckPulseDecay verifies the decay rate of a two-mode pulse against the
// analytic amplification-factor prediction.  A pulse built from cosine modes
// m=1 and m=3 must shrink exactly by g_1^k in the first mode and g_3^k in
// the second after k steps.  Wrong theta makes the measured decay disagree.
func CheckPulseDecay(op operator.Operator, k int) Outcome {
	name := "pulse decay matches theta=1/2"
	modes := []operator.ModeAmplitude{{Mode: 1, Amplitude: 1.0}, {Mode: 3, Amplitude: 0.5}}
	init := superposeModes(op, modes)
	final, err := op.AdvanceN(init, k, nil)
	if err != nil {
		return Outcome{Name: name, Pass: false, Detail: err.Error()}
	}
	expected, err := operator.EvolveCNClosedRod(op.Grid, op.Mu(), k, modes)
	if err != nil {
		return Outcome{Name: name, Pass: false, Detail: err.Error()}
	}
	worst := 0.0
	for i := range expected {
		d := math.Abs(expected[i] - final[i])
		if d > worst {
			worst = d
		}
	}
	ok := worst <= 1e-8
	detail := fmt.Sprintf("max |predicted - measured| = %.3g over %d steps", worst, k)
	if !ok {
		detail += " -- decay rate disagrees with the CN amplification factor"
	}
	return Outcome{Name: name, Pass: ok, Detail: detail}
}

// superposeModes builds the initial field A_1*cos(pi x/L) + A_3*cos(3 pi x/L)
// from the mode amplitudes.
func superposeModes(op operator.Operator, modes []operator.ModeAmplitude) []float64 {
	out := make([]float64, op.Grid.Nodes)
	for _, ma := range modes {
		mode, err := op.Grid.FourierMode(ma.Mode)
		if err != nil {
			continue
		}
		for i, v := range mode {
			out[i] += ma.Amplitude * v
		}
	}
	return out
}

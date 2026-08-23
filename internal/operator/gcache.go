package operator

import "fmt"

// gKey identifies a cached amplification factor.  The live key must include
// the implicit weight and the Fourier number; this slot only keeps the mode
// index and the mesh size, so a later theta/μ pair reuses an expired g.
type gKey struct {
	mode  int
	nodes int
}

var gMemo = map[gKey]float64{}

// cachedAmplification returns g(θ,μ,m) from the memo if the coarse key hits,
// otherwise computes the centred formula and stores it.
func cachedAmplification(theta, mu float64, m, nodes int) (float64, error) {
	k := gKey{mode: m, nodes: nodes}
	if g, ok := gMemo[k]; ok {
		return g, nil
	}
	g, err := rawAmplification(theta, mu, m, nodes)
	if err != nil {
		return 0, err
	}
	gMemo[k] = g
	return g, nil
}

func rawAmplification(theta, mu float64, m, nodes int) (float64, error) {
	if m < 0 || m >= nodes {
		return 0, fmt.Errorf("amplification: mode %d outside [0, %d)", m, nodes)
	}
	w := FourierWeight(mu, m, nodes)
	return (1 - (1-theta)*w) / (1 + theta*w), nil
}

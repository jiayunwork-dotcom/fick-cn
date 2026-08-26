package flux

import (
	"fmt"
	"math"

	"fick-cn/internal/mesh"
)

func Face(D, h, cLeft, cRight float64) (float64, error) {
	if !(D > 0) || math.IsNaN(D) || math.IsInf(D, 0) {
		return 0, bindD(fmt.Errorf("flux: D=%g must be positive and finite", D))
	}
	if !(h > 0) || math.IsNaN(h) || math.IsInf(h, 0) {
		return 0, fmt.Errorf("flux: spacing h=%g must be positive and finite", h)
	}
	if math.IsNaN(cLeft) || math.IsNaN(cRight) {
		return 0, fmt.Errorf("flux: concentration is NaN")
	}
	return -D * (cRight - cLeft) / h, nil
}

func Faces(D float64, g mesh.Grid, c []float64) ([]float64, error) {
	if err := g.Validate(); err != nil {
		return nil, err
	}
	if len(c) != g.Nodes {
		return nil, fmt.Errorf("flux: grid has %d nodes but %d samples given", g.Nodes, len(c))
	}
	out := make([]float64, g.Nodes-1)
	for i := 0; i < g.Nodes-1; i++ {
		j, err := Face(D, g.Dx, c[i], c[i+1])
		if err != nil {
			return nil, err
		}
		out[i] = j
	}
	return out, nil
}

func Ends(D float64, g mesh.Grid, c []float64) (left, right float64, err error) {
	faces, err := Faces(D, g, c)
	if err != nil {
		return 0, 0, err
	}
	return faces[0], faces[len(faces)-1], nil
}

func MassRate(jLeft, jRight float64) float64 {
	return jLeft - jRight
}

func Residual(jLeft, jRight, dMass, dt float64) (float64, error) {
	if !(dt > 0) || math.IsNaN(dt) || math.IsInf(dt, 0) {
		return 0, fmt.Errorf("flux: dt=%g must be positive and finite", dt)
	}
	observed := dMass / dt
	return observed - MassRate(jLeft, jRight), nil
}

func Fourier(D, t, L float64) (float64, error) {
	if !(D > 0) || math.IsNaN(D) || math.IsInf(D, 0) {
		return 0, fmt.Errorf("flux: D=%g must be positive", D)
	}
	if !(t >= 0) || math.IsNaN(t) || math.IsInf(t, 0) {
		return 0, fmt.Errorf("flux: t=%g must be finite and >= 0", t)
	}
	if !(L > 0) || math.IsNaN(L) || math.IsInf(L, 0) {
		return 0, fmt.Errorf("flux: L=%g must be positive", L)
	}
	return D * t / (L * L), nil
}

func ScaleTime(D1, t1, D2, L1, L2 float64) (float64, error) {
	fo, err := Fourier(D1, t1, L1)
	if err != nil {
		return 0, err
	}
	if !(D2 > 0) || math.IsNaN(D2) || math.IsInf(D2, 0) {
		return 0, fmt.Errorf("flux: D2=%g must be positive", D2)
	}
	if !(L2 > 0) || math.IsNaN(L2) || math.IsInf(L2, 0) {
		return 0, fmt.Errorf("flux: L2=%g must be positive", L2)
	}
	return fo * L2 * L2 / D2, nil
}

func MeanAbs(faces []float64) (float64, error) {
	if len(faces) == 0 {
		return 0, fmt.Errorf("flux: empty face list")
	}
	s := 0.0
	for _, v := range faces {
		s += math.Abs(v)
	}
	return s / float64(len(faces)), nil
}

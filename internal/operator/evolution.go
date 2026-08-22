package operator

import (
	"fmt"
	"math"

	"fick-cn/internal/mesh"
)

// ModeAmplitude pairs a Fourier mode index with its initial amplitude.  A
// superposition of these modes is the exact solution of the no-flux
// discrete diffusion equation, because every cosine mode is an eigenvector
// of the ghost-reflected Laplacian.
type ModeAmplitude struct {
	Mode      int
	Amplitude float64
}

// EvolveClosedRod predicts the field after k steps starting from a
// superposition of no-flux cosine modes.  Each mode m decays by its own
// amplification factor g_m raised to the number of steps:
//
//	c_k(x_i) = sum_m A_m * g_m^k * cos(m*pi*x_i/L)
//
// The comparison field is what a correctly weighted Crank–Nicolson solver
// must reproduce exactly.  A wrong theta (fully implicit or fully explicit)
// changes g_m and therefore the decay rate of every mode, making the
// predicted pulse decay visibly different.
func EvolveClosedRod(grid mesh.Grid, theta, mu float64, k int, modes []ModeAmplitude) ([]float64, error) {
	if k < 0 {
		return nil, fmt.Errorf("evolution: step count %d must be non-negative", k)
	}
	if err := grid.Validate(); err != nil {
		return nil, err
	}
	out := make([]float64, grid.Nodes)
	for _, ma := range modes {
		g, err := Amplification(theta, mu, ma.Mode, grid.Nodes)
		if err != nil {
			return nil, err
		}
		factor := math.Pow(g, float64(k))
		for i := 0; i < grid.Nodes; i++ {
			out[i] += ma.Amplitude * factor * math.Cos(float64(ma.Mode)*math.Pi*grid.Position(i)/grid.Length)
		}
	}
	return out, nil
}

// EvolveCNClosedRod is EvolveClosedRod fixed to the Crank–Nicolson weight.
func EvolveCNClosedRod(grid mesh.Grid, mu float64, k int, modes []ModeAmplitude) ([]float64, error) {
	return EvolveClosedRod(grid, ThetaCN, mu, k, modes)
}

// ModeNorm computes the l2 norm of a field, used to verify that the cosine
// basis spans the field consistently.
func ModeNorm(f []float64) float64 {
	sum := 0.0
	for _, v := range f {
		sum += v * v
	}
	return math.Sqrt(sum)
}

// CheckModeInvertibility verifies numerically that the cosine modes m=0..N-1
// are linearly independent on the grid, i.e. that the square matrix
// C[i][m] = cos(m*pi*x_i/L) is invertible.  It returns the largest absolute
// pivot encountered during a plain Gaussian elimination with partial
// pivoting.  A result close to zero would mean the basis is rank deficient.
func CheckModeInvertibility(grid mesh.Grid) float64 {
	n := grid.Nodes
	mat := make([][]float64, n)
	for i := 0; i < n; i++ {
		mat[i] = make([]float64, n)
		for m := 0; m < n; m++ {
			mat[i][m] = math.Cos(float64(m) * math.Pi * grid.Position(i) / grid.Length)
		}
	}
	minPivot := math.Inf(1)
	for col := 0; col < n; col++ {
		best := col
		for r := col + 1; r < n; r++ {
			if math.Abs(mat[r][col]) > math.Abs(mat[best][col]) {
				best = r
			}
		}
		mat[col], mat[best] = mat[best], mat[col]
		pivot := mat[col][col]
		if math.Abs(pivot) < math.Abs(minPivot) {
			minPivot = pivot
		}
		for r := col + 1; r < n; r++ {
			ratio := mat[r][col] / pivot
			for cc := col; cc < n; cc++ {
				mat[r][cc] -= ratio * mat[col][cc]
			}
		}
	}
	return math.Abs(minPivot)
}

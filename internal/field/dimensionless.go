package field

import (
	"math"
	"strconv"
)

// DiffusionLength returns sqrt(2*D*t), the characteristic length scale over
// which a diffusing spot spreads in one dimension.  Doubling the product D*t
// (e.g. by multiplying D by 4 and dividing t by 4) leaves the scale
// unchanged, which is the core of the D-t scaling identity:
//
//	c(x, t; D) = c(x, t/4; 4D)   for every x.
//
// The two runs share the same dimensionless profile at the same product D*t.
func DiffusionLength(D, t float64) float64 {
	return math.Sqrt(2 * D * t)
}

// DimensionlessTime returns D*t, the dimensionless diffusion time.
func DimensionlessTime(D, t float64) float64 {
	return D * t
}

// Coincide reports whether two fields are indistinguishable after rescaling
// the diffusion time by a factor k: runA used D and t, runB used k*D and
// t/k.  When both use the same grid and step schedule the profiles must
// agree to machine precision; the tolerance is kept tight so any real drift
// in how D enters the scheme is caught.
func Coincide(a, b Field, relTol float64) (bool, float64, error) {
	rel, err := RelativeDiff(a, b)
	if err != nil {
		return false, 0, err
	}
	return rel <= relTol, rel, nil
}

// ScalingReport describes a D-t scaling comparison between two runs.
type ScalingReport struct {
	BaseD       float64
	BaseT       float64
	ScaledD     float64
	ScaledT     float64
	ProductBefore float64
	ProductAfter  float64
	RelativeDiff float64
	Coincident   bool
}

// NewScalingReport summarises a cross-run comparison.
func NewScalingReport(baseD, baseT, scaledD, scaledT, relDiff float64, ok bool) ScalingReport {
	return ScalingReport{
		BaseD:         baseD,
		BaseT:         baseT,
		ScaledD:       scaledD,
		ScaledT:       scaledT,
		ProductBefore: DimensionlessTime(baseD, baseT),
		ProductAfter:  DimensionlessTime(scaledD, scaledT),
		RelativeDiff:  relDiff,
		Coincident:    ok,
	}
}

// Describe renders a scaling report as one line.
func (r ScalingReport) Describe() string {
	return "D*t product " + format(r.ProductBefore) + " vs " + format(r.ProductAfter) +
		", relative diff " + format(r.RelativeDiff) + ", coincident=" + boolStr(r.Coincident)
}

func format(v float64) string {
	return strconv.FormatFloat(v, 'g', 6, 64)
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

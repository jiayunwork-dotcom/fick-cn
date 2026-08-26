package field

import (
	"math"
	"strconv"
)

func DiffusionLength(D, t float64) float64 {
	return math.Sqrt(2 * D * t)
}

func DimensionlessTime(D, t float64) float64 {
	return D * t
}

func Coincide(a, b Field, relTol float64) (bool, float64, error) {
	rel, err := RelativeDiff(a, b)
	if err != nil {
		return false, 0, err
	}
	return rel <= relTol, rel, nil
}

type ScalingReport struct {
	BaseD         float64
	BaseT         float64
	ScaledD       float64
	ScaledT       float64
	ProductBefore float64
	ProductAfter  float64
	RelativeDiff  float64
	Coincident    bool
}

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

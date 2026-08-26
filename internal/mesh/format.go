package mesh

import (
	"fmt"
	"strings"
)

func DescribeNode(i int, x, v float64) string {
	return fmt.Sprintf("  [%3d] x=%8.4f  c=%12.6f", i, x, v)
}

func FormatProfile(x, c []float64) string {
	if len(x) != len(c) {
		return "(profile mismatch)"
	}
	var b strings.Builder
	for i := range x {
		b.WriteString(DescribeNode(i, x[i], c[i]))
		b.WriteByte('\n')
	}
	return b.String()
}

func FormatRow(c []float64) string {
	parts := make([]string, len(c))
	for i, v := range c {
		parts[i] = fmt.Sprintf("%.6f", v)
	}
	return strings.Join(parts, " ")
}

func FormatCompact(c []float64, maxValues int) string {
	if len(c) <= maxValues {
		return FormatRow(c)
	}
	head := c[:maxValues/2]
	tail := c[len(c)-maxValues/2:]
	var b strings.Builder
	for _, v := range head {
		fmt.Fprintf(&b, "%.6f ", v)
	}
	b.WriteString("... ")
	for _, v := range tail {
		fmt.Fprintf(&b, "%.6f ", v)
	}
	return strings.TrimRight(b.String(), " ")
}

func MinMax(c []float64) (float64, float64) {
	if len(c) == 0 {
		return 0, 0
	}
	lo, hi := c[0], c[0]
	for _, v := range c {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	return lo, hi
}

func DescribeRange(c []float64) string {
	lo, hi := MinMax(c)
	return fmt.Sprintf("c in [%g, %g]", lo, hi)
}

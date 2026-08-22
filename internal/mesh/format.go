package mesh

import (
	"fmt"
	"strings"
)

// DescribeNode writes a single line of a profile dump: the node index, its
// coordinate, and the sampled value.
func DescribeNode(i int, x, v float64) string {
	return fmt.Sprintf("  [%3d] x=%8.4f  c=%12.6f", i, x, v)
}

// FormatProfile renders every node of the rod as one line per node.  It is
// used by the CLI step and steady subcommands to expose the full field.
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

// FormatRow renders the whole field as a single space-separated line of
// values, useful for compact reports and for piping into plotting tools.
func FormatRow(c []float64) string {
	parts := make([]string, len(c))
	for i, v := range c {
		parts[i] = fmt.Sprintf("%.6f", v)
	}
	return strings.Join(parts, " ")
}

// FormatCompact renders the field as up to maxValues numbers; when the field
// is longer the middle is elided so the extremes stay visible.
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

// MinMax returns the minimum and maximum sampled value.
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

// DescribeRange returns a human-readable range summary of a field.
func DescribeRange(c []float64) string {
	lo, hi := MinMax(c)
	return fmt.Sprintf("c in [%g, %g]", lo, hi)
}

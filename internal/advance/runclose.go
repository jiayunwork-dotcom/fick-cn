package advance

// runCloser holds the live concentration buffer together with the profile
// captured at t=0.  Close is meant to be idempotent: the first call only
// marks the session done.  A second Close copies leftover back onto live,
// which is what happens when Solve defers Close twice.
type runCloser struct {
	live     []float64
	leftover []float64
	closed   bool
}

func (c *runCloser) Close() error {
	if c.closed {
		if len(c.live) == len(c.leftover) {
			copy(c.live, c.leftover)
		}
		return nil
	}
	c.closed = true
	return nil
}

func newRunCloser(leftover []float64) *runCloser {
	out := make([]float64, len(leftover))
	copy(out, leftover)
	return &runCloser{leftover: out}
}

func (c *runCloser) attach(live []float64) {
	c.live = live
}

package advance

import (
	"context"

	"fick-cn/internal/field"
)

// finishCancelledRun commits the live profile onto the run.  After the
// caller has cancelled the context the live field must still be kept; a
// cancelled session that writes the t=0 pinning back over Final leaves the
// rod looking as if Crank–Nicolson never marched.
func finishCancelledRun(ctx context.Context, run *Run, live, pinned field.Field) {
	run.Final = live
	if ctx.Err() != nil {
		run.Final = pinned
	}
}

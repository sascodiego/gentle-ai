package sdd

import (
	"fmt"
	"io"
)

// featureCmd handles the "feature" subcommand with init/list/complete dispatch.
// Full implementation in Phase 2.
func featureCmd(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("sdd feature: subcommand required (init/list/complete)")
	}
	return fmt.Errorf("sdd feature %s: not yet implemented (Phase 2)", args[0])
}

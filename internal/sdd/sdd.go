package sdd

import (
	"fmt"
	"io"
)

// Run is the sole public entry point for the sdd package.
// It dispatches subcommands based on args.
func Run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return fmt.Errorf("sdd: no subcommand provided")
	}
	switch args[0] {
	case "slug":
		return slugCmd(args[1:], stdout)
	case "current":
		return currentCmd(args[1:], stdout)
	case "feature":
		return featureCmd(args[1:], stdout)
	default:
		printUsage(stdout)
		return fmt.Errorf("sdd: unknown subcommand %q", args[0])
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, "Usage: gentle-ai sdd <subcommand> [options]\n\n")
	fmt.Fprintf(w, "Subcommands:\n")
	fmt.Fprintf(w, "  current          Show active change context\n")
	fmt.Fprintf(w, "  slug             Generate a branch slug\n")
	fmt.Fprintf(w, "  feature          Manage feature branches (init/list/complete)\n")
}

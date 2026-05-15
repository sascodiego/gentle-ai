package sdd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// currentCmd handles the "current" subcommand.
// Resolves the active SDD context and outputs structured information.
func currentCmd(args []string, stdout io.Writer) error {
	// Parse flags.
	jsonOutput := false
	branchOnly := false
	statusOnly := false
	for _, arg := range args {
		switch arg {
		case "--json":
			jsonOutput = true
		case "--branch":
			branchOnly = true
		case "--status":
			statusOnly = true
		}
	}

	ctx, err := ResolveContext()
	if err != nil {
		return err
	}

	// Handle single-field flags.
	if branchOnly {
		fmt.Fprintln(stdout, ctx.Branch)
		return nil
	}
	if statusOnly {
		fmt.Fprintln(stdout, ctx.Status)
		return nil
	}

	// Handle JSON output.
	if jsonOutput {
		type currentOutput struct {
			Change string `json:"change"`
			Branch string `json:"branch"`
			Path   string `json:"path"`
			Status string `json:"status"`
		}
		out := currentOutput{
			Change: ctx.Slug,
			Branch: ctx.Branch,
			Path:   ctx.ChangeDir,
			Status: string(ctx.Status),
		}
		data, err := json.Marshal(out)
		if err != nil {
			return fmt.Errorf("json marshal: %w", err)
		}
		fmt.Fprintln(stdout, string(data))
		return nil
	}

	// Default: human-readable output.
	var b strings.Builder
	fmt.Fprintf(&b, "change: %s\n", ctx.Slug)
	fmt.Fprintf(&b, "branch: %s\n", ctx.Branch)
	fmt.Fprintf(&b, "path: %s\n", ctx.ChangeDir)
	fmt.Fprintf(&b, "status: %s\n", ctx.Status)
	io.WriteString(stdout, b.String())

	return nil
}

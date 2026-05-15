package sdd

import (
	"fmt"
	"io"
	"strings"
)

// slugCmd handles the "slug" subcommand.
// Phase 2 will add full flag parsing; for now it's a minimal stub.
func slugCmd(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("slug: description argument required")
	}
	var slugType string
	var desc string
	for i := 0; i < len(args); i++ {
		if args[i] == "--type" && i+1 < len(args) {
			slugType = args[i+1]
			i++
		} else if !strings.HasPrefix(args[i], "--") {
			desc = args[i]
		}
	}
	if desc == "" {
		return fmt.Errorf("slug: description argument required")
	}
	opts := SlugOptions{Type: slugType}
	slug, err := Generate(desc, opts)
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, slug)
	return nil
}

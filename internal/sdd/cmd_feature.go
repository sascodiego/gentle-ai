package sdd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// featureCmd handles the "feature" subcommand with init/list/complete dispatch.
func featureCmd(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("sdd feature: subcommand required (init/list/complete)")
	}

	switch args[0] {
	case "init":
		return featureInitCmd(args[1:], stdout)
	case "list":
		return featureListCmd(args[1:], stdout)
	case "complete":
		return featureCompleteCmd(args[1:], stdout)
	default:
		return fmt.Errorf("sdd feature: unknown subcommand %q", args[0])
	}
}

// featureInitCmd creates a new feature branch, change directory, and active-change.yaml.
func featureInitCmd(args []string, stdout io.Writer) error {
	var slugType string
	var description string

	for i := 0; i < len(args); i++ {
		if args[i] == "--type" && i+1 < len(args) {
			slugType = args[i+1]
			i++
		} else if !strings.HasPrefix(args[i], "--") {
			description = args[i]
		}
	}

	if description == "" {
		return fmt.Errorf("sdd feature init: description argument required")
	}

	// Default type is "feat".
	if slugType == "" {
		slugType = "feat"
	}

	// Generate slug from description.
	opts := SlugOptions{Type: slugType}
	fullSlug, err := Generate(description, opts)
	if err != nil {
		return fmt.Errorf("sdd feature init: %w", err)
	}

	// Extract the slug part (after type/).
	slug := fullSlug
	if idx := strings.Index(fullSlug, "/"); idx >= 0 {
		slug = fullSlug[idx+1:]
	}

	branchName := slugType + "/" + slug

	// Create branch.
	if err := createBranch(branchName); err != nil {
		return fmt.Errorf("sdd feature init: %w", err)
	}

	// Create change directory.
	changeDir := filepath.Join(changesDir, slug)
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		return fmt.Errorf("sdd feature init: create change dir: %w", err)
	}

	// Write active-change.yaml.
	now := time.Now().UTC().Format(time.RFC3339)
	ac := &ActiveChange{
		ChangeName: slug,
		Type:       slugType,
		Branch:     branchName,
		Status:     string(PhaseExploring),
		CreatedAt:  now,
	}
	if err := WriteActiveChange(activeChangePath, ac); err != nil {
		return fmt.Errorf("sdd feature init: %w", err)
	}

	fmt.Fprintf(stdout, "Created:\n")
	fmt.Fprintf(stdout, "  branch: %s\n", branchName)
	fmt.Fprintf(stdout, "  directory: %s\n", changeDir)
	fmt.Fprintf(stdout, "  config: %s\n", activeChangePath)

	return nil
}

// featureListCmd enumerates openspec/changes/ and outputs slug + status per line.
func featureListCmd(args []string, stdout io.Writer) error {
	entries, err := os.ReadDir(changesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No changes directory = no changes.
		}
		return fmt.Errorf("sdd feature list: %w", err)
	}

	// Sort by name for deterministic output.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Exclude archive directory.
		if name == "archive" {
			continue
		}
		changeDir := filepath.Join(changesDir, name)
		status := DetectStatus(changeDir)
		fmt.Fprintf(stdout, "%s\t%s\n", name, status)
	}

	return nil
}

// featureCompleteCmd marks a feature as completed after verification.
func featureCompleteCmd(args []string, stdout io.Writer) error {
	ctx, err := ResolveContext()
	if err != nil {
		return fmt.Errorf("sdd feature complete: %w", err)
	}

	// Guard: status must be "verifying".
	if ctx.Status != PhaseVerifying {
		return fmt.Errorf("sdd feature complete: status must be verifying, current status is %q", ctx.Status)
	}

	// Update active-change.yaml to completed.
	ac := &ActiveChange{
		ChangeName: ctx.Slug,
		Type:       ctx.Type,
		Branch:     ctx.Branch,
		Status:     string(PhaseCompleted),
		CreatedAt:  "", // preserve — we don't have original, ReadActiveChange would give it
	}

	// Read original to preserve created_at.
	orig, err := ReadActiveChange(activeChangePath)
	if err == nil && orig.CreatedAt != "" {
		ac.CreatedAt = orig.CreatedAt
	}

	if err := WriteActiveChange(activeChangePath, ac); err != nil {
		return fmt.Errorf("sdd feature complete: %w", err)
	}

	fmt.Fprintf(stdout, "Marked %q as completed\n", ctx.Slug)
	return nil
}

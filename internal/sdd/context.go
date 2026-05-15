package sdd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ChangeContext holds resolved information about the active SDD change.
type ChangeContext struct {
	Slug      string // change slug (e.g., "sdd-branch-automation")
	Type      string // branch type prefix (e.g., "feat", "fix")
	Branch    string // full branch name (e.g., "feat/sdd-branch-automation")
	ChangeDir string // absolute path to openspec/changes/{slug}/
	Status    Phase  // detected lifecycle phase
}

// activeChangePath is the path to active-change.yaml.
// Variable seam for testing — production code uses the default.
var activeChangePath = "openspec/active-change.yaml"

// changesDir is the base directory for change artifacts.
// Variable seam for testing.
var changesDir = "openspec/changes"

// ResolveContext determines the active SDD change context.
//
// Resolution order:
// 1. Try git branch name → parse "{type}/{slug}" pattern
// 2. Fallback: read openspec/active-change.yaml → change_name
// 3. Neither resolves → error
func ResolveContext() (*ChangeContext, error) {
	// Strategy 1: Try git branch.
	branch, branchErr := currentBranch()
	if branchErr == nil && branch != "" {
		ctx, err := parseBranchContext(branch)
		if err == nil {
			return finalizeContext(ctx)
		}
		// Branch doesn't match pattern — fall through to yaml.
		_ = err
	}

	// Strategy 2: Try active-change.yaml.
	ac, err := ReadActiveChange(activeChangePath)
	if err == nil && ac.ChangeName != "" {
		ctx := &ChangeContext{
			Slug:   ac.ChangeName,
			Type:   ac.Type,
			Branch: branchOrFallback(branch, ac.Branch),
		}
		return finalizeContext(ctx)
	}

	// Strategy 3: Neither resolved.
	if branchErr != nil || branch == "" {
		return nil, fmt.Errorf("no active SDD context: not on a feature branch and no active-change.yaml found")
	}

	return nil, fmt.Errorf("no active SDD context: branch %q does not match type/slug pattern and no active-change.yaml found", branch)
}

// parseBranchContext parses a branch name in "type/slug" format.
func parseBranchContext(branch string) (*ChangeContext, error) {
	parts := strings.SplitN(branch, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("branch %q has no type/slash separator", branch)
	}
	slug := parts[1]
	typ := parts[0]

	// Validate: slug must not contain additional slashes.
	if strings.Contains(slug, "/") {
		return nil, fmt.Errorf("invalid branch pattern: %q (multiple slashes)", branch)
	}

	return &ChangeContext{
		Slug:   slug,
		Type:   typ,
		Branch: branch,
	}, nil
}

// finalizeContext fills in ChangeDir and Status for a context.
func finalizeContext(ctx *ChangeContext) (*ChangeContext, error) {
	ctx.ChangeDir = filepath.Join(changesDir, ctx.Slug)
	ctx.Status = DetectStatus(ctx.ChangeDir)
	return ctx, nil
}

// branchOrFallback returns the current branch if non-empty, otherwise the fallback.
func branchOrFallback(current, fallback string) string {
	if current != "" {
		return current
	}
	return fallback
}

// newActiveChangePath returns the absolute path for active-change.yaml
// relative to the given base directory.
func newActiveChangePath(baseDir string) string {
	return filepath.Join(baseDir, "openspec", "active-change.yaml")
}

// newChangesDir returns the absolute path for the changes directory
// relative to the given base directory.
func newChangesDir(baseDir string) string {
	return filepath.Join(baseDir, "openspec", "changes")
}

// ensureDir creates a directory and all parents if they don't exist.
func ensureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0o755)
	}
	return nil
}

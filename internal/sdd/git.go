package sdd

import (
	"fmt"
	"os/exec"
	"strings"
)

// Variable seams for testing — tests can override these without touching os/exec.
var (
	currentBranchFn = defaultCurrentBranch
	createBranchFn  = defaultCreateBranch
)

// currentBranch returns the name of the current git branch.
func currentBranch() (string, error) {
	return currentBranchFn()
}

// createBranch creates and checks out a new git branch with the given name.
func createBranch(name string) error {
	return createBranchFn(name)
}

func defaultCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch --show-current: %w", err)
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" {
		return "", fmt.Errorf("not on any branch (detached HEAD)")
	}
	return branch, nil
}

func defaultCreateBranch(name string) error {
	cmd := exec.Command("git", "checkout", "-b", name)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout -b %s: %w", name, err)
	}
	return nil
}

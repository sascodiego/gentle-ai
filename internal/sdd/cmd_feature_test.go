package sdd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureInit(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "init")

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"feature", "init", "Add branch automation"}, &buf)
	if err != nil {
		t.Fatalf("Run([feature init]) error: %v", err)
	}

	// Verify branch was created.
	branch, _ := currentBranch()
	if branch != "feat/add-branch-automation" {
		t.Errorf("branch = %q, want %q", branch, "feat/add-branch-automation")
	}

	// Verify directory was created.
	changeDir := filepath.Join(repoDir, "openspec", "changes", "add-branch-automation")
	info, err := os.Stat(changeDir)
	if err != nil {
		t.Fatalf("change dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("change path is not a directory")
	}

	// Verify yaml was created.
	yamlPath := filepath.Join(repoDir, "openspec", "active-change.yaml")
	ac, err := ReadActiveChange(yamlPath)
	if err != nil {
		t.Fatalf("ReadActiveChange error: %v", err)
	}
	if ac.ChangeName != "add-branch-automation" {
		t.Errorf("yaml change_name = %q, want %q", ac.ChangeName, "add-branch-automation")
	}
	if ac.Type != "feat" {
		t.Errorf("yaml type = %q, want %q", ac.Type, "feat")
	}
	if ac.Status != "exploring" {
		t.Errorf("yaml status = %q, want %q", ac.Status, "exploring")
	}

	// Verify output mentions created resources.
	output := buf.String()
	if !strings.Contains(output, "add-branch-automation") {
		t.Errorf("output missing slug, got:\n%s", output)
	}
}

func TestFeatureInitWithType(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "init")

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"feature", "init", "--type", "fix", "Fix login bug"}, &buf)
	if err != nil {
		t.Fatalf("Run([feature init --type fix]) error: %v", err)
	}

	branch, _ := currentBranch()
	if branch != "fix/fix-login-bug" {
		t.Errorf("branch = %q, want %q", branch, "fix/fix-login-bug")
	}

	yamlPath := filepath.Join(repoDir, "openspec", "active-change.yaml")
	ac, err := ReadActiveChange(yamlPath)
	if err != nil {
		t.Fatalf("ReadActiveChange error: %v", err)
	}
	if ac.Type != "fix" {
		t.Errorf("yaml type = %q, want %q", ac.Type, "fix")
	}
}

func TestFeatureInitNoDescription(t *testing.T) {
	var buf bytes.Buffer
	err := Run([]string{"feature", "init"}, &buf)
	if err == nil {
		t.Error("expected error with no description, got nil")
	}
}

func TestFeatureList(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "init")

	// Create some change directories.
	changesDir := filepath.Join(repoDir, "openspec", "changes")
	changeA := filepath.Join(changesDir, "change-alpha")
	changeB := filepath.Join(changesDir, "change-beta")
	if err := os.MkdirAll(changeA, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(changeB, 0o755); err != nil {
		t.Fatal(err)
	}
	// Give changeA a proposal so it shows as "proposed".
	os.WriteFile(filepath.Join(changeA, "proposal.md"), []byte("# P"), 0o644)

	// Create archive/ — should be excluded.
	if err := os.MkdirAll(filepath.Join(changesDir, "archive"), 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"feature", "list"}, &buf)
	if err != nil {
		t.Fatalf("Run([feature list]) error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "change-alpha") {
		t.Errorf("list missing 'change-alpha', got:\n%s", output)
	}
	if !strings.Contains(output, "change-beta") {
		t.Errorf("list missing 'change-beta', got:\n%s", output)
	}
	if strings.Contains(output, "archive") {
		t.Errorf("list should not contain 'archive', got:\n%s", output)
	}
	// Verify status is shown.
	if !strings.Contains(output, "proposed") {
		t.Errorf("list should show 'proposed' status for change-alpha, got:\n%s", output)
	}
}

func TestFeatureListEmpty(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "init")

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"feature", "list"}, &buf)
	if err != nil {
		t.Fatalf("Run([feature list]) error: %v", err)
	}
	if buf.String() != "" {
		t.Errorf("expected empty output for empty changes, got:\n%s", buf.String())
	}
}

func TestFeatureComplete(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "init")
	runGit(t, repoDir, "checkout", "-b", "feat/complete-test")

	// Set up change dir in verifying state.
	changesDir := filepath.Join(repoDir, "openspec", "changes", "complete-test")
	if err := os.MkdirAll(changesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create verify-report only — status will be "verifying" (not all core artifacts present).
	os.WriteFile(filepath.Join(changesDir, "verify-report.md"), []byte("# V"), 0o644)

	// Write active-change.yaml.
	yamlPath := filepath.Join(repoDir, "openspec", "active-change.yaml")
	ac := &ActiveChange{
		ChangeName: "complete-test",
		Type:       "feat",
		Branch:     "feat/complete-test",
		Status:     "verifying",
		CreatedAt:  "2025-01-01T00:00:00Z",
	}
	WriteActiveChange(yamlPath, ac)

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"feature", "complete"}, &buf)
	if err != nil {
		t.Fatalf("Run([feature complete]) error: %v", err)
	}

	// Verify yaml was updated to completed.
	updated, err := ReadActiveChange(yamlPath)
	if err != nil {
		t.Fatalf("ReadActiveChange after complete: %v", err)
	}
	if updated.Status != "completed" {
		t.Errorf("status after complete = %q, want %q", updated.Status, "completed")
	}
}

func TestFeatureCompleteNotVerifying(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "init")
	runGit(t, repoDir, "checkout", "-b", "feat/not-verifying")

	// Set up change dir in implementing state (tasks.md with check).
	changesDir := filepath.Join(repoDir, "openspec", "changes", "not-verifying")
	if err := os.MkdirAll(changesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(changesDir, "tasks.md"), []byte("# T\n- [x] Done"), 0o644)

	// Write active-change.yaml with implementing status.
	yamlPath := filepath.Join(repoDir, "openspec", "active-change.yaml")
	ac := &ActiveChange{
		ChangeName: "not-verifying",
		Type:       "feat",
		Branch:     "feat/not-verifying",
		Status:     "implementing",
		CreatedAt:  "2025-01-01T00:00:00Z",
	}
	WriteActiveChange(yamlPath, ac)

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"feature", "complete"}, &buf)
	if err == nil {
		t.Error("expected error when status is not verifying, got nil")
	}
	if !strings.Contains(err.Error(), "verifying") {
		t.Errorf("error should mention verifying, got: %v", err)
	}
}

func TestFeatureCompleteNoContext(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "init")

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"feature", "complete"}, &buf)
	if err == nil {
		t.Error("expected error with no active change, got nil")
	}
}

func TestFeatureNoSubcommand(t *testing.T) {
	var buf bytes.Buffer
	err := Run([]string{"feature"}, &buf)
	if err == nil {
		t.Error("expected error with no feature subcommand, got nil")
	}
}

func TestFeatureUnknownSubcommand(t *testing.T) {
	var buf bytes.Buffer
	err := Run([]string{"feature", "unknown"}, &buf)
	if err == nil {
		t.Error("expected error with unknown feature subcommand, got nil")
	}
}

func TestFeatureInitOutputFormat(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "init")

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"feature", "init", "My new feature"}, &buf)
	if err != nil {
		t.Fatalf("Run([feature init]) error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "branch:") {
		t.Errorf("output should contain 'branch:', got:\n%s", output)
	}
	if !strings.Contains(output, "directory:") {
		t.Errorf("output should contain 'directory:', got:\n%s", output)
	}
	if !strings.Contains(output, "config:") {
		t.Errorf("output should contain 'config:', got:\n%s", output)
	}
}

func TestFeatureListSortedAlphabetically(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "init")

	changesDir := filepath.Join(repoDir, "openspec", "changes")
	// Create in non-alphabetical order.
	for _, name := range []string{"zulu-change", "alpha-change", "mid-change"} {
		if err := os.MkdirAll(filepath.Join(changesDir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"feature", "list"}, &buf)
	if err != nil {
		t.Fatalf("Run([feature list]) error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), lines)
	}
	if !strings.HasPrefix(lines[0], "alpha-change") {
		t.Errorf("first line should start with 'alpha-change', got: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "mid-change") {
		t.Errorf("second line should start with 'mid-change', got: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "zulu-change") {
		t.Errorf("third line should start with 'zulu-change', got: %q", lines[2])
	}
}

func TestFeatureListExcludesNonDirectories(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "init")

	changesDir := filepath.Join(repoDir, "openspec", "changes")
	if err := os.MkdirAll(changesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create a file (not directory) in changes.
	os.WriteFile(filepath.Join(changesDir, "README.md"), []byte("# Changes"), 0o644)
	// Create one real directory.
	if err := os.MkdirAll(filepath.Join(changesDir, "real-change"), 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"feature", "list"}, &buf)
	if err != nil {
		t.Fatalf("Run([feature list]) error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "README") {
		t.Errorf("list should exclude files, got:\n%s", output)
	}
	if !strings.Contains(output, "real-change") {
		t.Errorf("list should include directories, got:\n%s", output)
	}
}

// Ensure tests compile with exec import.
var _ = exec.Command

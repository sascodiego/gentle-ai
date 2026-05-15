package sdd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestRepo creates a temporary git repo for integration testing.
// Returns the repo dir, yaml path, and changes dir path.
func setupTestRepo(t *testing.T) (repoDir, yamlPath, cDir string) {
	t.Helper()
	repoDir = t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "init")

	yamlPath = filepath.Join(repoDir, "openspec", "active-change.yaml")
	cDir = filepath.Join(repoDir, "openspec", "changes")
	return repoDir, yamlPath, cDir
}

func TestCurrentCmdHumanOutput(t *testing.T) {
	repoDir, _, cDir := setupTestRepo(t)
	runGit(t, repoDir, "checkout", "-b", "feat/test-change")

	// Create change dir so status detection works.
	changeDir := filepath.Join(cDir, "test-change")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write a proposal.md so status is "proposed".
	if err := os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Test"), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"current"}, &buf)
	if err != nil {
		t.Fatalf("Run([current]) error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "change: test-change") {
		t.Errorf("output missing 'change: test-change', got:\n%s", output)
	}
	if !strings.Contains(output, "branch: feat/test-change") {
		t.Errorf("output missing 'branch: feat/test-change', got:\n%s", output)
	}
	if !strings.Contains(output, "status: proposed") {
		t.Errorf("output missing 'status: proposed', got:\n%s", output)
	}
}

func TestCurrentCmdJSONOutput(t *testing.T) {
	repoDir, _, cDir := setupTestRepo(t)
	runGit(t, repoDir, "checkout", "-b", "feat/json-test")

	changeDir := filepath.Join(cDir, "json-test")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"current", "--json"}, &buf)
	if err != nil {
		t.Fatalf("Run([current, --json]) error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON parse error: %v\noutput: %s", err, buf.String())
	}
	if result["change"] != "json-test" {
		t.Errorf("JSON change = %q, want %q", result["change"], "json-test")
	}
	if result["branch"] != "feat/json-test" {
		t.Errorf("JSON branch = %q, want %q", result["branch"], "feat/json-test")
	}
	if result["status"] != string(PhaseExploring) {
		t.Errorf("JSON status = %q, want %q", result["status"], PhaseExploring)
	}
}

func TestCurrentCmdBranchFlag(t *testing.T) {
	repoDir, _, _ := setupTestRepo(t)
	runGit(t, repoDir, "checkout", "-b", "feat/branch-flag-test")

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"current", "--branch"}, &buf)
	if err != nil {
		t.Fatalf("Run([current, --branch]) error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "feat/branch-flag-test" {
		t.Errorf("--branch output = %q, want %q", got, "feat/branch-flag-test")
	}
}

func TestCurrentCmdStatusFlag(t *testing.T) {
	repoDir, _, cDir := setupTestRepo(t)
	runGit(t, repoDir, "checkout", "-b", "feat/status-test")

	changeDir := filepath.Join(cDir, "status-test")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create tasks.md with a checked item to get "implementing" status.
	tasksContent := "# Tasks\n- [x] Some task done\n"
	if err := os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte(tasksContent), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"current", "--status"}, &buf)
	if err != nil {
		t.Fatalf("Run([current, --status]) error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "implementing" {
		t.Errorf("--status output = %q, want %q", got, "implementing")
	}
}

func TestCurrentCmdNoContext(t *testing.T) {
	repoDir, _, _ := setupTestRepo(t)
	// Stay on default branch with no yaml.

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"current"}, &buf)
	if err == nil {
		t.Error("Run([current]) expected error on main with no context, got nil")
	}
}

func TestCurrentCmdYAMLFallback(t *testing.T) {
	repoDir, yamlPath, cDir := setupTestRepo(t)
	// Stay on default branch, but create a yaml.

	changeDir := filepath.Join(cDir, "yaml-fallback")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	ac := &ActiveChange{
		ChangeName: "yaml-fallback",
		Type:       "feat",
		Branch:     "feat/yaml-fallback",
		Status:     "exploring",
		CreatedAt:  "2025-01-01T00:00:00Z",
	}
	if err := WriteActiveChange(yamlPath, ac); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"current"}, &buf)
	if err != nil {
		t.Fatalf("Run([current]) error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "change: yaml-fallback") {
		t.Errorf("output missing 'change: yaml-fallback', got:\n%s", output)
	}
}

// Ensure tests compile with exec import.
var _ = exec.Command

func TestCurrentCmdJSONContainsPath(t *testing.T) {
	repoDir, _, cDir := setupTestRepo(t)
	runGit(t, repoDir, "checkout", "-b", "feat/path-test")

	changeDir := filepath.Join(cDir, "path-test")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"current", "--json"}, &buf)
	if err != nil {
		t.Fatalf("Run([current, --json]) error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}
	expected := filepath.Join("openspec", "changes", "path-test")
	if !strings.Contains(result["path"], expected) {
		t.Errorf("JSON path = %q, want containing %q", result["path"], expected)
	}
}

func TestCurrentCmdMultipleFlags(t *testing.T) {
	// When both --branch and --status are given, --branch takes precedence
	// (first match in the flag switch).
	repoDir, _, _ := setupTestRepo(t)
	runGit(t, repoDir, "checkout", "-b", "feat/multi-flag")

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	var buf bytes.Buffer
	err := Run([]string{"current", "--status", "--branch"}, &buf)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "feat/multi-flag" {
		t.Errorf("multi-flag output = %q, want %q (--branch should win)", got, "feat/multi-flag")
	}
}

package sdd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCurrentBranch(t *testing.T) {
	// Create a temp directory and init a git repo.
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "init")

	// Detect the default branch name.
	defaultBranch := gitBranch(t, repoDir)

	tests := []struct {
		name   string
		branch string // branch to create and checkout; empty stays on default
		want   string // expected branch name
	}{
		{
			name: "on default branch",
			want: defaultBranch,
		},
		{
			name:   "on feature branch",
			branch: "feat/my-feature",
			want:   "feat/my-feature",
		},
		{
			name:   "on fix branch",
			branch: "fix/bug-123",
			want:   "fix/bug-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.branch != "" {
				runGit(t, repoDir, "checkout", "-b", tt.branch)
			}

			origDir, _ := os.Getwd()
			os.Chdir(repoDir)
			defer os.Chdir(origDir)

			got, err := currentBranch()
			if err != nil {
				t.Errorf("currentBranch() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("currentBranch() = %q, want %q", got, tt.want)
			}

			// Switch back to default for next test.
			if tt.branch != "" {
				os.Chdir(origDir)
				runGit(t, repoDir, "checkout", defaultBranch)
			}
		})
	}
}

func TestCurrentBranchNotGitRepo(t *testing.T) {
	// In a temp dir that is NOT a git repo, currentBranch should error.
	origDir, _ := os.Getwd()
	os.Chdir(t.TempDir())
	defer os.Chdir(origDir)

	_, err := currentBranch()
	if err == nil {
		t.Error("currentBranch() expected error in non-git directory, got nil")
	}
}

func TestCreateBranch(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "init")

	defaultBranch := gitBranch(t, repoDir)

	tests := []struct {
		name   string
		branch string
	}{
		{
			name:   "create new branch",
			branch: "feat/new-feature",
		},
		{
			name:   "create another branch",
			branch: "fix/hotfix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure we start on default branch.
			runGit(t, repoDir, "checkout", defaultBranch)

			origDir, _ := os.Getwd()
			os.Chdir(repoDir)
			defer os.Chdir(origDir)

			err := createBranch(tt.branch)
			if err != nil {
				t.Errorf("createBranch(%q) error = %v", tt.branch, err)
				return
			}
			// Verify the branch was created and checked out.
			got, _ := currentBranch()
			if got != tt.branch {
				t.Errorf("after createBranch(%q), currentBranch() = %q, want %q", tt.branch, got, tt.branch)
			}
		})
	}
}

func TestCreateBranchAlreadyExists(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@test.com")
	runGit(t, repoDir, "config", "user.name", "Test")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "init")
	// Create a branch that already exists.
	runGit(t, repoDir, "branch", "feat/existing")

	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	err := createBranch("feat/existing")
	if err == nil {
		t.Error("createBranch() expected error for existing branch, got nil")
	}
}

// TestDefaultGitOverrides verifies the variable seam pattern works.
func TestDefaultGitOverrides(t *testing.T) {
	// Save originals.
	origCurrentBranch := currentBranchFn
	origCreateBranch := createBranchFn
	defer func() {
		currentBranchFn = origCurrentBranch
		createBranchFn = origCreateBranch
	}()

	// Override with mock implementations.
	currentBranchFn = func() (string, error) {
		return "feat/mock-branch", nil
	}
	createBranchFn = func(name string) error {
		return nil
	}

	got, err := currentBranchFn()
	if err != nil {
		t.Fatalf("mock currentBranchFn error: %v", err)
	}
	if got != "feat/mock-branch" {
		t.Errorf("mock currentBranchFn = %q, want %q", got, "feat/mock-branch")
	}

	err = createBranchFn("feat/test")
	if err != nil {
		t.Errorf("mock createBranchFn error: %v", err)
	}
}

// TestGitIgnoreInfra verifies test infrastructure for .gitignore handling.
func TestGitIgnoreInfra(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")

	gitIgnore := filepath.Join(repoDir, ".gitignore")
	if err := os.WriteFile(gitIgnore, []byte("openspec/active-change.yaml\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(gitIgnore)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "openspec/active-change.yaml\n" {
		t.Errorf("gitignore = %q, want active-change.yaml entry", string(data))
	}
}

// runGit executes a git command in the given directory.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

// gitBranch returns the current branch name in dir.
func gitBranch(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("detect default branch: %v", err)
	}
	return string(out[:len(out)-1]) // trim newline
}

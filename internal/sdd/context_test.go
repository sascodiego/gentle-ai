package sdd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveContext(t *testing.T) {
	// Save and restore variable seams.
	origBranchFn := currentBranchFn
	defer func() { currentBranchFn = origBranchFn }()

	tests := []struct {
		name        string
		branch      string // mock branch name (empty = error)
		branchErr   error  // if set, mock returns this error
		yamlContent string // content for active-change.yaml (empty = no file)
		wantSlug    string
		wantType    string
		wantBranch  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "branch resolves feat/slug",
			branch:     "feat/my-change",
			wantSlug:   "my-change",
			wantType:   "feat",
			wantBranch: "feat/my-change",
		},
		{
			name:       "branch resolves fix/slug",
			branch:     "fix/bug-fix",
			wantSlug:   "bug-fix",
			wantType:   "fix",
			wantBranch: "fix/bug-fix",
		},
		{
			name:       "branch resolves chore/slug",
			branch:     "chore/cleanup",
			wantSlug:   "cleanup",
			wantType:   "chore",
			wantBranch: "chore/cleanup",
		},
		{
			name:        "branch with multiple slashes errors",
			branch:      "feat/my/change",
			wantErr:     true,
			errContains: "no active SDD context",
		},
		{
			name:        "branch without slash falls through",
			branch:      "some-branch",
			yamlContent: "change_name: yaml-change\ntype: feat\nbranch: feat/yaml-change\nstatus: exploring\ncreated_at: 2025-01-01T00:00:00Z\n",
			wantSlug:    "yaml-change",
			wantType:    "feat",
			wantBranch:  "some-branch", // branch is what git reports
		},
		{
			name:        "yaml fallback when no branch",
			branchErr:   errNoBranch,
			yamlContent: "change_name: yaml-fallback\ntype: fix\nbranch: fix/yaml-fallback\nstatus: proposed\ncreated_at: 2025-01-01T00:00:00Z\n",
			wantSlug:    "yaml-fallback",
			wantType:    "fix",
			wantBranch:  "fix/yaml-fallback",
		},
		{
			name:        "no branch no yaml errors",
			branchErr:   errNoBranch,
			wantErr:     true,
			errContains: "no active SDD context",
		},
		{
			name:        "yaml missing change_name errors",
			branchErr:   errNoBranch,
			yamlContent: "type: feat\nstatus: exploring\n",
			wantErr:     true,
			errContains: "no active SDD context",
		},
		{
			name:       "docs type branch",
			branch:     "docs/readme-update",
			wantSlug:   "readme-update",
			wantType:   "docs",
			wantBranch: "docs/readme-update",
		},
		{
			name:       "revert type branch",
			branch:     "revert/bad-commit",
			wantSlug:   "bad-commit",
			wantBranch: "revert/bad-commit",
			wantType:   "revert",
		},
		{
			name:        "branch with no slash and no yaml",
			branch:      "main",
			wantErr:     true,
			errContains: "no active SDD context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up temp dir for openspec/active-change.yaml.
			tmpDir := t.TempDir()
			yamlPath := filepath.Join(tmpDir, "openspec", "active-change.yaml")
			if tt.yamlContent != "" {
				if err := os.MkdirAll(filepath.Dir(yamlPath), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(yamlPath, []byte(tt.yamlContent), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			// Mock currentBranch.
			if tt.branchErr != nil {
				currentBranchFn = func() (string, error) {
					return "", tt.branchErr
				}
			} else {
				currentBranchFn = func() (string, error) {
					return tt.branch, nil
				}
			}

			// Save/restore activeChangePath override.
			origPath := activeChangePath
			activeChangePath = yamlPath
			defer func() { activeChangePath = origPath }()

			// Save/restore changesDir override.
			origChangesDir := changesDir
			changesDir = filepath.Join(tmpDir, "openspec", "changes")
			defer func() { changesDir = origChangesDir }()

			ctx, err := ResolveContext()
			if tt.wantErr {
				if err == nil {
					t.Error("ResolveContext() expected error, got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ResolveContext() error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveContext() unexpected error: %v", err)
			}
			if ctx.Slug != tt.wantSlug {
				t.Errorf("Slug = %q, want %q", ctx.Slug, tt.wantSlug)
			}
			if ctx.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", ctx.Type, tt.wantType)
			}
			if ctx.Branch != tt.wantBranch {
				t.Errorf("Branch = %q, want %q", ctx.Branch, tt.wantBranch)
			}
		})
	}
}

func TestResolveContextChangeDir(t *testing.T) {
	origBranchFn := currentBranchFn
	defer func() { currentBranchFn = origBranchFn }()

	currentBranchFn = func() (string, error) {
		return "feat/my-change", nil
	}

	tmpDir := t.TempDir()
	origChangesDir := changesDir
	changesDir = filepath.Join(tmpDir, "openspec", "changes")
	defer func() { changesDir = origChangesDir }()

	ctx, err := ResolveContext()
	if err != nil {
		t.Fatalf("ResolveContext() error: %v", err)
	}
	expected := filepath.Join(tmpDir, "openspec", "changes", "my-change")
	if ctx.ChangeDir != expected {
		t.Errorf("ChangeDir = %q, want %q", ctx.ChangeDir, expected)
	}
}

func TestResolveContextStatus(t *testing.T) {
	origBranchFn := currentBranchFn
	defer func() { currentBranchFn = origBranchFn }()

	currentBranchFn = func() (string, error) {
		return "feat/my-change", nil
	}

	tmpDir := t.TempDir()
	origChangesDir := changesDir
	changesDir = filepath.Join(tmpDir, "openspec", "changes")
	defer func() { changesDir = origChangesDir }()

	// Create change dir with proposal.md to get "proposed" status.
	changeDir := filepath.Join(tmpDir, "openspec", "changes", "my-change")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, err := ResolveContext()
	if err != nil {
		t.Fatalf("ResolveContext() error: %v", err)
	}
	if ctx.Status != PhaseProposed {
		t.Errorf("Status = %q, want %q", ctx.Status, PhaseProposed)
	}
}

// errNoBranch is a sentinel error for "not on any branch".
type testError string

func (e testError) Error() string { return string(e) }

var errNoBranch = testError("not on any branch")

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

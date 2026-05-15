package sdd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectStatus(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(dir string) // create artifacts in dir
		expected Phase
	}{
		{
			name:     "empty directory is exploring",
			setup:    func(dir string) {},
			expected: PhaseExploring,
		},
		{
			name: "exploration.md exists",
			setup: func(dir string) {
				mustWriteFile(t, filepath.Join(dir, "exploration.md"), "# Exploration")
			},
			expected: PhaseExploring,
		},
		{
			name: "proposal.md exists",
			setup: func(dir string) {
				mustWriteFile(t, filepath.Join(dir, "proposal.md"), "# Proposal")
			},
			expected: PhaseProposed,
		},
		{
			name: "specs directory with spec file",
			setup: func(dir string) {
				mustMkdir(t, filepath.Join(dir, "specs"))
				mustWriteFile(t, filepath.Join(dir, "specs", "change.md"), "# Spec")
			},
			expected: PhaseSpecd,
		},
		{
			name: "design.md exists",
			setup: func(dir string) {
				mustWriteFile(t, filepath.Join(dir, "design.md"), "# Design")
			},
			expected: PhaseDesigned,
		},
		{
			name: "proposal + specs + design → designed",
			setup: func(dir string) {
				mustWriteFile(t, filepath.Join(dir, "proposal.md"), "# Proposal")
				mustMkdir(t, filepath.Join(dir, "specs"))
				mustWriteFile(t, filepath.Join(dir, "design.md"), "# Design")
			},
			expected: PhaseDesigned,
		},
		{
			name: "tasks.md exists",
			setup: func(dir string) {
				mustWriteFile(t, filepath.Join(dir, "tasks.md"), "# Tasks")
			},
			expected: PhaseTasked,
		},
		{
			name: "tasks.md with checked items → implementing",
			setup: func(dir string) {
				mustWriteFile(t, filepath.Join(dir, "tasks.md"), "# Tasks\n\n- [x] Task 1 done\n- [ ] Task 2 pending\n")
			},
			expected: PhaseImplementing,
		},
		{
			name: "apply-progress.md exists → implementing",
			setup: func(dir string) {
				mustWriteFile(t, filepath.Join(dir, "apply-progress.md"), "progress")
			},
			expected: PhaseImplementing,
		},
		{
			name: "apply-progress directory exists → implementing",
			setup: func(dir string) {
				mustMkdir(t, filepath.Join(dir, "apply-progress"))
			},
			expected: PhaseImplementing,
		},
		{
			name: "verify-report.md exists → verifying",
			setup: func(dir string) {
				mustWriteFile(t, filepath.Join(dir, "verify-report.md"), "# Verify")
			},
			expected: PhaseVerifying,
		},
		{
			name: "verify-report directory exists → verifying",
			setup: func(dir string) {
				mustMkdir(t, filepath.Join(dir, "verify-report"))
			},
			expected: PhaseVerifying,
		},
		{
			name: "archive-report.md exists → archived",
			setup: func(dir string) {
				mustWriteFile(t, filepath.Join(dir, "archive-report.md"), "# Archive")
			},
			expected: PhaseArchived,
		},
		{
			name: "all artifacts plus verify-report → completed",
			setup: func(dir string) {
				mustWriteFile(t, filepath.Join(dir, "proposal.md"), "# Proposal")
				mustMkdir(t, filepath.Join(dir, "specs"))
				mustWriteFile(t, filepath.Join(dir, "specs", "change.md"), "# Spec")
				mustWriteFile(t, filepath.Join(dir, "design.md"), "# Design")
				mustWriteFile(t, filepath.Join(dir, "tasks.md"), "# Tasks\n- [x] Done\n")
				mustWriteFile(t, filepath.Join(dir, "verify-report.md"), "# Verify")
			},
			expected: PhaseCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(dir)
			got := DetectStatus(dir)
			if got != tt.expected {
				t.Errorf("DetectStatus(%q) = %q, want %q", dir, got, tt.expected)
			}
		})
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

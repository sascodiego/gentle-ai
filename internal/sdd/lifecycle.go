package sdd

import (
	"os"
	"path/filepath"
	"strings"
)

// Phase represents a lifecycle stage of an SDD change.
type Phase string

const (
	PhaseExploring    Phase = "exploring"
	PhaseProposed     Phase = "proposed"
	PhaseSpecd        Phase = "spec'd"
	PhaseDesigned     Phase = "designed"
	PhaseTasked       Phase = "tasked"
	PhaseImplementing Phase = "implementing"
	PhaseVerifying    Phase = "verifying"
	PhaseCompleted    Phase = "completed"
	PhaseArchived     Phase = "archived"
)

// DetectStatus determines the lifecycle phase of a change by checking
// artifact presence in the change directory.
// Detection rules checked in order (first match wins):
// 1. archive-report.md exists → archived
// 2. All core artifacts + verify-report.md → completed
// 3. verify-report.md or verify-report/ → verifying
// 4. tasks.md has checked items OR apply-progress exists → implementing
// 5. tasks.md exists → tasked
// 6. design.md exists → designed
// 7. specs/ has content → spec'd
// 8. proposal.md exists → proposed
// 9. Directory exists → exploring
func DetectStatus(changeDir string) Phase {
	// Check archived first (highest priority).
	if fileExists(filepath.Join(changeDir, "archive-report.md")) {
		return PhaseArchived
	}

	// Check completed: all core artifacts present + verify-report.
	hasProposal := fileExists(filepath.Join(changeDir, "proposal.md"))
	hasDesign := fileExists(filepath.Join(changeDir, "design.md"))
	hasTasks := fileExists(filepath.Join(changeDir, "tasks.md"))
	hasSpecs := dirHasContent(filepath.Join(changeDir, "specs"))
	hasVerifyReport := fileExists(filepath.Join(changeDir, "verify-report.md")) || dirExists(filepath.Join(changeDir, "verify-report"))

	if hasProposal && hasDesign && hasTasks && hasSpecs && hasVerifyReport {
		return PhaseCompleted
	}

	// Check verifying.
	if hasVerifyReport {
		return PhaseVerifying
	}

	// Check implementing: tasks.md has checked items OR apply-progress exists.
	if hasTasks {
		data, err := os.ReadFile(filepath.Join(changeDir, "tasks.md"))
		if err == nil && strings.Contains(string(data), "- [x]") {
			return PhaseImplementing
		}
	}
	if fileExists(filepath.Join(changeDir, "apply-progress.md")) || dirExists(filepath.Join(changeDir, "apply-progress")) {
		return PhaseImplementing
	}

	// Check tasked.
	if hasTasks {
		return PhaseTasked
	}

	// Check designed.
	if hasDesign {
		return PhaseDesigned
	}

	// Check spec'd.
	if hasSpecs {
		return PhaseSpecd
	}

	// Check proposed.
	if hasProposal {
		return PhaseProposed
	}

	// Default: exploring.
	return PhaseExploring
}

// fileExists reports whether path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// dirExists reports whether path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// dirHasContent reports whether dir exists and contains at least one file.
func dirHasContent(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return len(entries) > 0
}

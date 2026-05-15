package sdd

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteAndReadActiveChange(t *testing.T) {
	original := &ActiveChange{
		ChangeName: "sdd-branch-automation",
		Type:       "feat",
		Branch:     "feat/sdd-branch-automation",
		Status:     "implementing",
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "active-change.yaml")

	// Write
	if err := WriteActiveChange(path, original); err != nil {
		t.Fatalf("WriteActiveChange() error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("active-change.yaml was not created")
	}

	// Read back
	got, err := ReadActiveChange(path)
	if err != nil {
		t.Fatalf("ReadActiveChange() error: %v", err)
	}

	if got.ChangeName != original.ChangeName {
		t.Errorf("ChangeName = %q, want %q", got.ChangeName, original.ChangeName)
	}
	if got.Type != original.Type {
		t.Errorf("Type = %q, want %q", got.Type, original.Type)
	}
	if got.Branch != original.Branch {
		t.Errorf("Branch = %q, want %q", got.Branch, original.Branch)
	}
	if got.Status != original.Status {
		t.Errorf("Status = %q, want %q", got.Status, original.Status)
	}
	if got.CreatedAt != original.CreatedAt {
		t.Errorf("CreatedAt = %q, want %q", got.CreatedAt, original.CreatedAt)
	}
}

func TestReadActiveChangeMissingFile(t *testing.T) {
	_, err := ReadActiveChange("/nonexistent/active-change.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReadActiveChangeMalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "active-change.yaml")
	if err := os.WriteFile(path, []byte("not valid yaml content\nthat is missing fields"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadActiveChange(path)
	if err == nil {
		t.Error("expected error for malformed YAML")
	}
}

func TestWriteActiveChangeCreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "nested", "active-change.yaml")

	ac := &ActiveChange{
		ChangeName: "test-change",
		Type:       "fix",
		Branch:     "fix/test-change",
		Status:     "proposed",
		CreatedAt:  "2026-01-01T00:00:00Z",
	}

	if err := WriteActiveChange(path, ac); err != nil {
		t.Fatalf("WriteActiveChange() error: %v", err)
	}

	got, err := ReadActiveChange(path)
	if err != nil {
		t.Fatalf("ReadActiveChange() error: %v", err)
	}
	if got.ChangeName != "test-change" {
		t.Errorf("ChangeName = %q, want %q", got.ChangeName, "test-change")
	}
}

func TestRoundTripMultipleTypes(t *testing.T) {
	types := []string{"feat", "fix", "chore", "docs", "refactor", "test"}
	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "active-change.yaml")

			ac := &ActiveChange{
				ChangeName: "my-change",
				Type:       typ,
				Branch:     typ + "/my-change",
				Status:     "exploring",
				CreatedAt:  "2026-01-15T12:00:00Z",
			}

			if err := WriteActiveChange(path, ac); err != nil {
				t.Fatalf("WriteActiveChange() error: %v", err)
			}

			got, err := ReadActiveChange(path)
			if err != nil {
				t.Fatalf("ReadActiveChange() error: %v", err)
			}
			if got.Type != typ {
				t.Errorf("Type = %q, want %q", got.Type, typ)
			}
		})
	}
}

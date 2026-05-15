package sdd

import (
	"bytes"
	"strings"
	"testing"
)

func TestSlugCmdBasic(t *testing.T) {
	var buf bytes.Buffer
	err := Run([]string{"slug", "Add user login page"}, &buf)
	if err != nil {
		t.Fatalf("Run([slug]) error: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "add-user-login-page" {
		t.Errorf("slug basic = %q, want %q", got, "add-user-login-page")
	}
}

func TestSlugCmdWithType(t *testing.T) {
	var buf bytes.Buffer
	err := Run([]string{"slug", "--type", "feat", "Add user login page"}, &buf)
	if err != nil {
		t.Fatalf("Run([slug, --type feat]) error: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "feat/add-user-login-page" {
		t.Errorf("slug with type = %q, want %q", got, "feat/add-user-login-page")
	}
}

func TestSlugCmdStopWords(t *testing.T) {
	var buf bytes.Buffer
	err := Run([]string{"slug", "Fix the error in the user model"}, &buf)
	if err != nil {
		t.Fatalf("Run([slug]) error: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "fix-error-user-model" {
		t.Errorf("slug stop words = %q, want %q", got, "fix-error-user-model")
	}
}

func TestSlugCmdNumeric(t *testing.T) {
	var buf bytes.Buffer
	err := Run([]string{"slug", "Add 2fa support"}, &buf)
	if err != nil {
		t.Fatalf("Run([slug]) error: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "add-2fa-support" {
		t.Errorf("slug numeric = %q, want %q", got, "add-2fa-support")
	}
}

func TestSlugCmdNoDescription(t *testing.T) {
	var buf bytes.Buffer
	err := Run([]string{"slug"}, &buf)
	if err == nil {
		t.Error("Run([slug]) expected error with no description, got nil")
	}
}

func TestSlugCmdEmptyDescription(t *testing.T) {
	var buf bytes.Buffer
	err := Run([]string{"slug", ""}, &buf)
	if err == nil {
		t.Error("Run([slug, '']) expected error, got nil")
	}
}

func TestSlugCmdTypeFlagOnly(t *testing.T) {
	var buf bytes.Buffer
	err := Run([]string{"slug", "--type", "feat"}, &buf)
	if err == nil {
		t.Error("Run([slug, --type, feat]) expected error with no description, got nil")
	}
}

package sdd

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		opts    SlugOptions
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			desc: "Add user login page",
			opts: SlugOptions{},
			want: "add-user-login-page",
		},
		{
			name: "stop words removed",
			desc: "Fix the error in the user model",
			opts: SlugOptions{},
			want: "fix-error-user-model",
		},
		{
			name: "with type prefix",
			desc: "Add user login page",
			opts: SlugOptions{Type: "feat"},
			want: "feat/add-user-login-page",
		},
		{
			name: "with fix type prefix",
			desc: "Handle null pointer in handler",
			opts: SlugOptions{Type: "fix"},
			want: "fix/handle-null-pointer-handler",
		},
		{
			name: "numeric words preserved",
			desc: "Add 2fa support",
			opts: SlugOptions{},
			want: "add-2fa-support",
		},
		{
			name: "all stop words fallback",
			desc: "to be or not to be",
			opts: SlugOptions{},
			// All words are stop words → fallback to all original words → first 4: [to, be, or, not]
			want: "to-be-or-not",
		},
		{
			name: "all stop words with type",
			desc: "to be or not to be",
			opts: SlugOptions{Type: "feat"},
			want: "feat/to-be-or-not",
		},
		{
			name: "special characters stripped",
			desc: "Add user/login & registration!",
			opts: SlugOptions{},
			want: "add-user-login-registration",
		},
		{
			name:    "empty description",
			desc:    "",
			opts:    SlugOptions{},
			wantErr: true,
		},
		{
			name: "single word",
			desc: "Authentication",
			opts: SlugOptions{},
			want: "authentication",
		},
		{
			name: "many words truncated to four",
			desc: "Add a new feature for the user to be able to login with their email",
			opts: SlugOptions{},
			want: "add-new-feature-user",
		},
		{
			name: "short words filtered after stop words",
			desc: "go to the rust conference",
			opts: SlugOptions{},
			// "go" (2 chars), "to" (stop word), "the" (stop word), "rust" (4), "conference" (10)
			// After stop words: [go, rust, conference]
			// After <3 filter: [rust, conference] (go is 2 chars)
			// First 4: [rust, conference]
			want: "rust-conference",
		},
		{
			name: "consecutive hyphens collapsed",
			desc: "fix -- the / bug",
			opts: SlugOptions{},
			// After stop words ("the"): ["fix", "--", "/", "bug"]
			// After <3 filter: ["fix", "bug"] (-- is 2 chars, / is 1 char)
			// Join: "fix-bug"
			want: "fix-bug",
		},
		{
			name: "244 byte boundary exact",
			desc: strings.Repeat("abcdefghij ", 50),
			opts: SlugOptions{MaxBytes: 50},
			want: "abcdefghij-abcdefghij-abcdefghij-abcdefghij",
		},
		{
			name: "default MaxBytes is 244",
			desc: strings.Repeat("word ", 100),
			opts: SlugOptions{},
			want: "word-word-word-word",
		},
		{
			name: "type prefix not counted toward byte limit",
			desc: strings.Repeat("word ", 100),
			opts: SlugOptions{Type: "feat", MaxBytes: 20},
			want: "feat/word-word-word-word",
		},
		{
			name: "truncation at hyphen boundary",
			desc: "aa bb cc dd ee ff",
			opts: SlugOptions{MaxBytes: 10},
			// All words survive (none stop, all >= 3 except "aa","bb","cc","dd","ee","ff" — 2 chars each)
			// After stop words: all remain
			// After <3 filter: all removed! Fallback to raw: [aa, bb, cc, dd, ee, ff]
			// First 4: [aa, bb, cc, dd] → "aa-bb-cc-dd" = 11 bytes > 10
			// Truncate at hyphen: "aa-bb-cc" = 8 bytes
			want: "aa-bb-cc",
		},
		{
			name:    "only punctuation input",
			desc:    "!@#$%^&*()",
			opts:    SlugOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Generate(tt.desc, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Generate(%q, %+v) = %q, want error", tt.desc, tt.opts, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Generate(%q, %+v) error: %v", tt.desc, tt.opts, err)
			}
			if got != tt.want {
				t.Errorf("Generate(%q, %+v) = %q, want %q", tt.desc, tt.opts, got, tt.want)
			}
		})
	}
}

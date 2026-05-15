package sdd

import (
	"fmt"
	"regexp"
	"strings"
)

// SlugOptions configures slug generation.
type SlugOptions struct {
	Type     string // feat|fix|chore|docs|refactor|revert — prepended as {type}/ prefix
	MaxBytes int    // default 244, applied to slug part (type prefix doesn't count toward limit)
}

// stopWords are removed from slug candidates.
// Matches the spec algorithm requirement list.
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true,
	"and": true, "or": true, "but": true,
	"in": true, "on": true, "at": true,
	"to": true, "for": true, "of": true,
	"with": true, "by": true, "from": true,
	"is": true, "it": true, "that": true, "this": true, "which": true,
	"be": true, "as": true, "was": true, "were": true, "been": true,
	"not": true, "no": true,
}

var nonSlugChar = regexp.MustCompile(`[^a-z0-9-]`)
var multiHyphen = regexp.MustCompile(`-{2,}`)

const defaultMaxBytes = 244

// Generate produces a slug from a description string.
//
// Algorithm: lowercase → split words → remove stop-words → filter <3 chars →
// first 4 meaningful words → join with "-" → sanitize (only a-z, 0-9, -) →
// collapse consecutive hyphens → truncate to MaxBytes at hyphen boundary.
// If all words are removed (stop-words + <3 filter), falls back to original words.
func Generate(desc string, opts SlugOptions) (string, error) {
	if strings.TrimSpace(desc) == "" {
		return "", fmt.Errorf("description must not be empty")
	}

	maxBytes := opts.MaxBytes
	if maxBytes <= 0 {
		maxBytes = defaultMaxBytes
	}

	words := strings.Fields(strings.ToLower(desc))

	// Collect meaningful words: not stop-words AND length >= 3.
	var meaningful []string
	for _, w := range words {
		if stopWords[w] {
			continue
		}
		if len(w) < 3 {
			continue
		}
		meaningful = append(meaningful, w)
	}

	// Fallback: if all words were filtered, use all original words.
	if len(meaningful) == 0 {
		meaningful = words
	}

	// Take first 4.
	if len(meaningful) > 4 {
		meaningful = meaningful[:4]
	}

	// Join with hyphens and sanitize.
	slug := strings.Join(meaningful, "-")
	slug = nonSlugChar.ReplaceAllString(slug, "-")
	slug = multiHyphen.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if slug == "" {
		return "", fmt.Errorf("slug produced no valid characters from %q", desc)
	}

	// Truncate slug part to MaxBytes at hyphen boundary.
	slug = truncateAtHyphen(slug, maxBytes)

	// Prepend type prefix if provided (prefix doesn't count toward byte limit).
	if opts.Type != "" {
		slug = opts.Type + "/" + slug
	}

	return slug, nil
}

// truncateAtHyphen truncates s to at most maxBytes, breaking at the last hyphen
// that keeps the result within the limit.
func truncateAtHyphen(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	// Find the last hyphen within the byte limit.
	lastHyphen := -1
	for i := 0; i < len(s) && i < maxBytes; i++ {
		if s[i] == '-' {
			lastHyphen = i
		}
	}
	if lastHyphen > 0 {
		return s[:lastHyphen]
	}
	// No hyphen found — hard truncate.
	return s[:maxBytes]
}

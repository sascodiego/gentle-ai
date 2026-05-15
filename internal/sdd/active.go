package sdd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ActiveChange represents the contents of active-change.yaml.
type ActiveChange struct {
	ChangeName string
	Type       string
	Branch     string
	Status     string
	CreatedAt  string
}

// ReadActiveChange reads and parses an active-change.yaml file.
func ReadActiveChange(path string) (*ActiveChange, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read active-change: %w", err)
	}
	defer f.Close()

	ac := &ActiveChange{}
	scanner := bufio.NewScanner(f)
	found := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := parseKV(line)
		if !ok {
			continue
		}
		switch key {
		case "change_name":
			ac.ChangeName = value
			found++
		case "type":
			ac.Type = value
			found++
		case "branch":
			ac.Branch = value
			found++
		case "status":
			ac.Status = value
			found++
		case "created_at":
			ac.CreatedAt = value
			found++
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read active-change: %w", err)
	}
	if found == 0 {
		return nil, fmt.Errorf("read active-change: no valid fields found in %s", path)
	}
	return ac, nil
}

// WriteActiveChange writes an ActiveChange to a YAML file.
// Creates parent directories if needed.
func WriteActiveChange(path string, ac *ActiveChange) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("write active-change: %w", err)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("change_name: %s\n", ac.ChangeName))
	b.WriteString(fmt.Sprintf("type: %s\n", ac.Type))
	b.WriteString(fmt.Sprintf("branch: %s\n", ac.Branch))
	b.WriteString(fmt.Sprintf("status: %s\n", ac.Status))
	b.WriteString(fmt.Sprintf("created_at: %s\n", ac.CreatedAt))

	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write active-change: %w", err)
	}
	return nil
}

// parseKV splits a "key: value" line.
func parseKV(line string) (key, value string, ok bool) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:idx])
	value = strings.TrimSpace(line[idx+1:])
	if key == "" {
		return "", "", false
	}
	return key, value, true
}

package licensebootstrap

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ReadEnvValue(path, key string) (string, error) {
	lines, err := readLines(path)
	if err != nil {
		return "", err
	}
	value := ""
	for _, line := range lines {
		if candidate, ok := activeValue(line, key); ok {
			value = candidate
		}
	}
	return value, nil
}

func SetEnvValue(path, key, value string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	return writeLines(path, setLine(lines, key, value))
}

func WriteElection(path, key, value string, unsetKeys ...string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	for _, unset := range unsetKeys {
		lines = unsetLines(lines, unset)
	}
	return writeLines(path, setLine(lines, key, value))
}

func setLine(lines []string, key, value string) []string {
	entry := key + "=" + value
	for i, line := range lines {
		if _, ok := activeValue(line, key); ok {
			lines[i] = entry
			return unsetLinesAfter(lines, key, i)
		}
	}
	for i, line := range lines {
		if isCommentedTemplate(line, key) {
			lines[i] = entry
			return lines
		}
	}
	return append(lines, entry)
}

func unsetLinesAfter(lines []string, key string, keep int) []string {
	kept := lines[:0]
	for i, line := range lines {
		if _, ok := activeValue(line, key); ok && i != keep {
			continue
		}
		kept = append(kept, line)
	}
	return kept
}

func unsetLines(lines []string, key string) []string {
	kept := lines[:0]
	for _, line := range lines {
		if _, ok := activeValue(line, key); ok {
			continue
		}
		kept = append(kept, line)
	}
	return kept
}

func UnsetEnvValue(path, key string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return nil
	}
	return writeLines(path, unsetLines(lines, key))
}

func activeValue(line, key string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") {
		return "", false
	}
	trimmed = strings.TrimPrefix(trimmed, "export ")
	name, value, found := strings.Cut(trimmed, "=")
	if !found || strings.TrimSpace(name) != key {
		return "", false
	}
	value = strings.TrimSpace(value)
	if len(value) >= 2 && strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		value = value[1 : len(value)-1]
	}
	return value, true
}

func isCommentedTemplate(line, key string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "#") {
		return false
	}
	trimmed = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
	name, _, found := strings.Cut(trimmed, "=")
	return found && strings.TrimSpace(name) == key
}

func ensureRegularFile(path string) (os.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("refusing to use %s: not a regular file", path)
	}
	return info, nil
}

func readLines(path string) ([]string, error) {
	expected, err := ensureRegularFile(path)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if expected == nil || !os.SameFile(expected, opened) {
		return nil, fmt.Errorf("refusing to use %s: file changed while being read", path)
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(path string, lines []string) error {
	if _, err := ensureRegularFile(path); err != nil {
		return err
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".env.*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

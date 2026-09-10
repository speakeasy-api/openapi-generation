package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeriveInstallationURL(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		lang         string
		repoURL      string
		subDirectory string
		expected     string
	}{
		"python with subdirectory": {
			lang:         "python",
			repoURL:      "https://github.com/org/repo",
			subDirectory: "python/sdk",
			expected:     "https://github.com/org/repo.git#subdirectory=python/sdk",
		},
		"python without subdirectory": {
			lang:     "python",
			repoURL:  "https://github.com/org/repo",
			expected: "https://github.com/org/repo.git",
		},
		"python repoURL already has .git": {
			lang:         "python",
			repoURL:      "https://github.com/org/repo.git",
			subDirectory: "sdk",
			expected:     "https://github.com/org/repo.git#subdirectory=sdk",
		},
		"python trailing slash on subdirectory": {
			lang:         "python",
			repoURL:      "https://github.com/org/repo",
			subDirectory: "python/sdk/",
			expected:     "https://github.com/org/repo.git#subdirectory=python/sdk",
		},
		"python dot subdirectory treated as empty": {
			lang:         "python",
			repoURL:      "https://github.com/org/repo",
			subDirectory: ".",
			expected:     "https://github.com/org/repo.git",
		},
		"python whitespace subdirectory treated as empty": {
			lang:         "python",
			repoURL:      "https://github.com/org/repo",
			subDirectory: "   ",
			expected:     "https://github.com/org/repo.git",
		},
		"typescript with subdirectory": {
			lang:         "typescript",
			repoURL:      "https://github.com/org/repo.git",
			subDirectory: "ts/sdk",
			expected:     "https://github.com/org/repo",
		},
		"typescript without subdirectory": {
			lang:     "typescript",
			repoURL:  "https://github.com/org/repo",
			expected: "https://github.com/org/repo",
		},
		"ruby with subdirectory": {
			lang:         "ruby",
			repoURL:      "https://github.com/org/repo.git",
			subDirectory: "ruby/sdk",
			expected:     "https://github.com/org/repo -d ruby/sdk",
		},
		"ruby without subdirectory": {
			lang:     "ruby",
			repoURL:  "https://github.com/org/repo",
			expected: "https://github.com/org/repo",
		},
		"php with subdirectory ignored": {
			lang:         "php",
			repoURL:      "https://github.com/org/repo",
			subDirectory: "php/sdk",
			expected:     "https://github.com/org/repo.git",
		},
		"php repoURL already has .git": {
			lang:     "php",
			repoURL:  "https://github.com/org/repo.git",
			expected: "https://github.com/org/repo.git",
		},
		"default language trims .git": {
			lang:     "go",
			repoURL:  "https://github.com/org/repo.git",
			expected: "https://github.com/org/repo",
		},
		"default language without .git": {
			lang:     "csharp",
			repoURL:  "https://github.com/org/repo",
			expected: "https://github.com/org/repo",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := deriveInstallationURL(tc.lang, tc.repoURL, tc.subDirectory)
			assert.Equal(t, tc.expected, got)
		})
	}
}

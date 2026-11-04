package main

import "testing"

func TestExtractMinVersion(t *testing.T) {
	tests := []struct {
		constraint string
		want       string
	}{
		{"^3.25.0", "3.25.0"},
		{"~3.25.0", "3.25.0"},
		{"~> 1.73.2", "1.73.2"},
		{"~>1.73.2", "1.73.2"},
		{">=8.2", "8.2"},
		{"> 8.2", "8.2"},
		{"3.25.0", "3.25.0"},
		{"v3.25.0", "3.25.0"},
		{"^3.25.0 || ^4.0.0", "3.25.0"},
		{"", ""},
	}

	for _, tt := range tests {
		if got := extractMinVersion(tt.constraint); got != tt.want {
			t.Errorf("extractMinVersion(%q) = %q, want %q", tt.constraint, got, tt.want)
		}
	}
}

func TestIsIgnoredVersions(t *testing.T) {
	ignoreList := []IgnoreEntry{
		{ID: "GHSA-1", Package: "zod", Versions: []string{"3.25.0", "3.25.65"}},
		{ID: "GHSA-2", Package: "qs"},
	}

	tests := []struct {
		id, pkg, version string
		want             bool
	}{
		{"GHSA-1", "zod", "3.25.0", true},
		{"GHSA-1", "zod", "3.25.65", true},
		{"GHSA-1", "zod", "4.2.0", false},
		{"GHSA-1", "zod", "", false},
		{"GHSA-1", "other", "3.25.0", false},
		{"GHSA-2", "qs", "", true},
		{"GHSA-2", "qs", "6.14.0", true},
		{"GHSA-3", "qs", "", false},
	}

	for _, tt := range tests {
		if got := isIgnored(tt.id, tt.pkg, tt.version, ignoreList); got != tt.want {
			t.Errorf("isIgnored(%q, %q, %q) = %v, want %v", tt.id, tt.pkg, tt.version, got, tt.want)
		}
	}
}

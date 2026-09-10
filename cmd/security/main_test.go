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

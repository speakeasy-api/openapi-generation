package ast

import "testing"

func TestAreEquivalentRequirements(t *testing.T) {
	tests := []struct {
		name string
		a    []SecurityRequirement
		b    []SecurityRequirement
		want bool
	}{
		{
			name: "identical single scheme",
			a:    []SecurityRequirement{{"apiKey"}},
			b:    []SecurityRequirement{{"apiKey"}},
			want: true,
		},
		{
			name: "different schemes",
			a:    []SecurityRequirement{{"apiKey"}},
			b:    []SecurityRequirement{{"oauth2"}},
			want: false,
		},
		{
			name: "AND order does not matter",
			a:    []SecurityRequirement{{"apiKey", "oauth2"}},
			b:    []SecurityRequirement{{"oauth2", "apiKey"}},
			want: true,
		},
		{
			name: "OR order does matter",
			a:    []SecurityRequirement{{"apiKey"}, {"oauth2"}},
			b:    []SecurityRequirement{{"oauth2"}, {"apiKey"}},
			want: false,
		},
		{
			name: "different lengths",
			a:    []SecurityRequirement{{"apiKey"}, {"oauth2"}},
			b:    []SecurityRequirement{{"apiKey"}},
			want: false,
		},
		{
			name: "both empty",
			a:    []SecurityRequirement{},
			b:    []SecurityRequirement{},
			want: true,
		},
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: true,
		},
		{
			name: "duplicate empty reqs ignored",
			a:    []SecurityRequirement{{"apiKey"}, {"oauth2"}, {}, {}},
			b:    []SecurityRequirement{{"apiKey"}, {"oauth2"}, {}},
			want: true,
		},
		{
			name: "optionality does matter",
			a:    []SecurityRequirement{{"apiKey"}, {}},
			b:    []SecurityRequirement{{"apiKey"}},
			want: false,
		},
		{
			name: "mixed requirements",
			a:    []SecurityRequirement{{"apiKey", "oauth2"}, {"bearerAuth"}},
			b:    []SecurityRequirement{{"apiKey", "oauth2"}, {"bearerAuth"}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AreEquivalentRequirements(tt.a, tt.b); got != tt.want {
				t.Errorf("AreEquivalentRequirements() = %v, want %v", got, tt.want)
			}
		})
	}
}

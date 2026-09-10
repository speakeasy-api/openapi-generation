package contenttypes_test

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/contenttypes"
	"github.com/stretchr/testify/assert"
)

func BenchmarkIsJSON(b *testing.B) {
	testCases := []string{
		"application/json",
		"application/json; charset=utf-8",
		"text/json",
		"application/vnd.api+json",
		"text/plain",
		"application/xml",
	}
	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				contenttypes.IsJSON(tc)
			}
		})
	}
}

func BenchmarkIsXML(b *testing.B) {
	testCases := []string{
		"application/xml",
		"application/xml; charset=utf-8",
		"text/xml",
		"application/vnd.api+xml",
		"text/plain",
		"application/json",
	}
	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				contenttypes.IsXML(tc)
			}
		})
	}
}

func BenchmarkIsYAML(b *testing.B) {
	testCases := []string{
		"application/yaml",
		"application/x-yaml",
		"text/yaml",
		"text/yml",
		"application/vnd.api+yaml",
		"text/plain",
		"application/json",
	}
	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				contenttypes.IsYAML(tc)
			}
		})
	}
}

func BenchmarkIsCSV(b *testing.B) {
	testCases := []string{
		"application/csv",
		"text/csv",
		"text/csv; charset=utf-8",
		"text/plain",
		"application/json",
	}
	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				contenttypes.IsCSV(tc)
			}
		})
	}
}

func BenchmarkIsMultipart(b *testing.B) {
	testCases := []string{
		"multipart/form-data",
		"multipart/mixed",
		"multipart/related",
		"text/plain",
		"application/json",
	}
	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				contenttypes.IsMultipart(tc)
			}
		})
	}
}

func BenchmarkIsURLEncoded(b *testing.B) {
	testCases := []string{
		"application/x-www-form-urlencoded",
		"application/x-www-form-urlencoded; charset=utf-8",
		"text/plain",
		"application/json",
	}
	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				contenttypes.IsURLEncoded(tc)
			}
		})
	}
}

func BenchmarkIsTextPlain(b *testing.B) {
	testCases := []string{
		"text/plain",
		"text/plain+custom",
		"text/html",
		"application/json",
	}
	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				contenttypes.IsTextPlain(tc)
			}
		})
	}
}

func BenchmarkIsEventStream(b *testing.B) {
	testCases := []string{
		"text/event-stream",
		"text/event-stream+custom",
		"text/plain",
		"application/json",
	}
	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				contenttypes.IsEventStream(tc)
			}
		})
	}
}

func BenchmarkIsJsonL(b *testing.B) {
	testCases := []string{
		"application/jsonl",
		"text/jsonl",
		"application/x-ndjson",
		"text/x-ndjson",
		"application/json",
		"text/plain",
	}
	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				contenttypes.IsJsonL(tc)
			}
		})
	}
}

func BenchmarkIsOctetStream(b *testing.B) {
	testCases := []string{
		"application/octet-stream",
		"application/octet-stream; charset=utf-8",
		"text/plain",
		"application/json",
	}
	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				contenttypes.IsOctetStream(tc)
			}
		})
	}
}

func TestIsJsonSeq(t *testing.T) {
	t.Parallel()

	tests := []struct {
		contentType string
		expected    bool
	}{
		{"application/json-seq", true},
		{"application/json-seq; charset=utf-8", true},
		{"application/vnd.api+json-seq", true},
		{"application/json", false},
		{"application/jsonl", false},
		{"text/plain", false},
		{"text/event-stream", false},
	}
	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, contenttypes.IsJsonSeq(tt.contentType))
		})
	}
}

func TestIsJSON_ExcludesSequentialTypes(t *testing.T) {
	t.Parallel()

	// json-seq and jsonl should NOT match IsJSON
	assert.False(t, contenttypes.IsJSON("application/json-seq"))
	assert.False(t, contenttypes.IsJSON("application/jsonl"))
	assert.False(t, contenttypes.IsJSON("application/x-ndjson"))

	// Regular JSON should still match
	assert.True(t, contenttypes.IsJSON("application/json"))
	assert.True(t, contenttypes.IsJSON("text/json"))
	assert.True(t, contenttypes.IsJSON("application/vnd.api+json"))
}

func TestSortAcceptTypes(t *testing.T) {
	type args struct {
		acceptTypes []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Sorts accept types",
			args: args{
				acceptTypes: []string{
					"application/json",
					"application/*",
					"application/yaml",
					"*/*",
					"application/csv",
					"text/json; charset=utf-8",
					"text/*",
					"text/plain",
					"text/csv",
					"application/csv; charset=utf-8",
					"multipart/form-data",
					"text/json",
				},
			},
			want: []string{
				"text/json; charset=utf-8",
				"application/json",
				"text/json",
				"application/yaml",
				"application/csv; charset=utf-8",
				"application/csv",
				"text/csv",
				"text/plain",
				"multipart/form-data",
				"text/*",
				"application/*",
				"*/*",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contenttypes.SortAcceptTypes(tt.args.acceptTypes)
			assert.Equal(t, tt.want, got)
		})
	}
}

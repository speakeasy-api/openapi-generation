package sanitization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SanitizeName_Success(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "succeeds with valid name",
			args: args{
				name: "some_name",
			},
			want: "some_name",
		},
		{
			name: "succeeds with name with spaces",
			args: args{
				name: "some name",
			},
			want: "some_name",
		},
		{
			name: "succeeds with name with emojis",
			args: args{
				name: "🚀 some_name",
			},
			want: "some_name",
		},
		{
			name: "succeeds handling multiple emojis",
			args: args{
				name: "🚀 some👩🏼‍❤️‍💋_name👩🏼‍❤️‍💋‍👨🏼",
			},
			want: "some_name_",
		},
		{
			name: "succeeds with diacritics",
			args: args{
				name: "À some_name",
			},
			want: "A_some_name",
		},
		{
			name: "succeeds with non-ascii characters in prefix",
			args: args{
				name: "≠some_name",
			},
			want: "some_name",
		},
		{
			name: "succeeds with ascii symbols in prefix",
			args: args{
				name: "<>some_name",
			},
			want: "LessThan_GreaterThan_some_name",
		},
		{
			name: "succeeds in not repeating underscores",
			args: args{
				name: "some[_name",
			},
			want: "some_name",
		},
		{
			name: "succeeds replacing prefix numbers",
			args: args{
				name: "1some_name",
			},
			want: "onesome_name",
		},
		{
			name: "succeeds replacing prefix numbers with following symbols",
			args: args{
				name: "1>some_name",
			},
			want: "oneGreaterThan_some_name",
		},
		{
			name: "succeeds replacing multiple prefix numbers with following symbols",
			args: args{
				name: "10>some_name",
			},
			want: "tenGreaterThan_some_name",
		},
		{
			name: "succeeds handling trailing numbers",
			args: args{
				name: "some_name10",
			},
			want: "some_name10",
		},
		{
			name: "succeeds with single underscores",
			args: args{
				name: "some_string_with_underscores",
			},
			want: "some_string_with_underscores",
		},
		{
			name: "succeeds with double underscores",
			args: args{
				name: "some__string_with_underscores",
			},
			want: "some_string_with_underscores",
		},
		{
			name: "succeeds with many underscores",
			args: args{
				name: "some___string__with_underscores__",
			},
			want: "some_string_with_underscores_",
		},
		{
			name: "succeeds with leading and trailing underscores",
			args: args{
				name: "_operation_with_leading_and_trailing_underscores_",
			},
			want: "_operation_with_leading_and_trailing_underscores_",
		},
		{
			name: "succeeds with multiple leading and trailing underscores",
			args: args{
				name: "__operation_with_multiple_leading_and_trailing_underscores__",
			},
			want: "_operation_with_multiple_leading_and_trailing_underscores_",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeName(tt.args.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizeMediaType_New(t *testing.T) {
	tests := []struct {
		mediaType string
		expected  string
	}{
		{
			mediaType: "*/*",
			expected:  "",
		},
		{
			mediaType: "application/x-www-form-urlencoded",
			expected:  "FormEncoded",
		},
		{
			mediaType: "multipart/form-data",
			expected:  "FormEncoded",
		},
		{
			mediaType: "application/json",
			expected:  "",
		},
		{
			mediaType: "application/vnd.api+json",
			expected:  "",
		},
		{
			mediaType: "application/json-seq",
			expected:  "",
		},
		{
			mediaType: "application/x-ndjson",
			expected:  "",
		},
		{
			mediaType: "application/ld+json",
			expected:  "",
		},
		{
			mediaType: "application/problem+json",
			expected:  "",
		},
		{
			mediaType: "image/png",
			expected:  "ImagePng",
		},
		{
			mediaType: "text/xml",
			expected:  "XML",
		},
		{
			mediaType: "image/*",
			expected:  "Image",
		},
		{
			mediaType: "application/xml",
			expected:  "XML",
		},
		{
			mediaType: "application/xml",
			expected:  "XML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.mediaType, func(t *testing.T) {
			got := HumanizeMediaType(tt.mediaType)
			if got != tt.expected {
				t.Errorf("SanitizeMediaType(%q) = %q, want %q", tt.mediaType, got, tt.expected)
			}
		})
	}
}

func TestSanitizeStatusCode_New(t *testing.T) {
	tests := []struct {
		statusCode string
		expected   string
	}{
		{
			statusCode: "default",
			expected:   "",
		},
		{
			statusCode: "200",
			expected:   "",
		},
		{
			statusCode: "1xx",
			expected:   "Informational",
		},
		{
			statusCode: "3XX",
			expected:   "Redirect",
		},
		{
			statusCode: "4xx",
			expected:   "ClientError",
		},
		{
			statusCode: "5XX",
			expected:   "ServerError",
		},
		{
			statusCode: "404",
			expected:   "NotFound",
		},
		{
			statusCode: "401",
			expected:   "Unauthorized",
		},
		{
			statusCode: "500",
			expected:   "InternalServerError",
		},
		{
			statusCode: "304",
			expected:   "NotModified",
		},
		{
			statusCode: "foo",
			expected:   "",
		},
		{
			statusCode: "  400 ",
			expected:   "BadRequest",
		},
		{
			statusCode: "2XX",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.statusCode, func(t *testing.T) {
			got := HumanizeStatusCode(tt.statusCode)
			if got != tt.expected {
				t.Errorf("SanitizeStatusCode(%q) = %q, want %q", tt.statusCode, got, tt.expected)
			}
		})
	}
}

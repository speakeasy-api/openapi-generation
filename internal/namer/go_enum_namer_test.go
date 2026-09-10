package namer

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/stretchr/testify/assert"
)

func TestGoEnumNamerEnumNames(t *testing.T) {
	t.Parallel()

	n := NewGoEnumNamer()

	tests := []struct {
		name     string
		typeDef  *ast.TypeDef
		expected []string
	}{
		{
			name:     "nil_enum_returns_nil",
			typeDef:  &ast.TypeDef{Name: "Status"},
			expected: nil,
		},
		// Values path
		{
			name: "values_simple",
			typeDef: &ast.TypeDef{
				Name: "Status",
				Enum: &ast.Enum{Values: []string{"active", "inactive"}},
			},
			expected: []string{"statusactive", "statusinactive"},
		},
		{
			name: "values_empty_string_becomes_unknown",
			typeDef: &ast.TypeDef{
				Name: "Status",
				Enum: &ast.Enum{Values: []string{""}},
			},
			expected: []string{"statusunknown"},
		},
		{
			name: "values_whitespace_only_becomes_unknown",
			typeDef: &ast.TypeDef{
				Name: "Status",
				Enum: &ast.Enum{Values: []string{"  "}},
			},
			expected: []string{"statusunknown"},
		},
		{
			name: "values_special_chars_sanitized",
			typeDef: &ast.TypeDef{
				Name: "Status",
				Enum: &ast.Enum{Values: []string{"in-progress"}},
			},
			expected: []string{"statusinprogress"},
		},
		{
			name: "values_disambiguation_upper_lower",
			typeDef: &ast.TypeDef{
				Name: "Status",
				Enum: &ast.Enum{Values: []string{"ACTIVE", "active"}},
			},
			expected: []string{"statusactiveupper", "statusactivelower"},
		},
		{
			name: "values_disambiguation_upper_mixed",
			typeDef: &ast.TypeDef{
				Name: "Status",
				Enum: &ast.Enum{Values: []string{"ACTIVE", "Active"}},
			},
			expected: []string{"statusactiveupper", "statusactivemixed"},
		},
		{
			name: "values_disambiguation_three_way",
			typeDef: &ast.TypeDef{
				Name: "Mode",
				Enum: &ast.Enum{Values: []string{"FAST", "fast", "Fast"}},
			},
			expected: []string{"modefastupper", "modefastlower", "modefastmixed"},
		},
		// Names path
		{
			name: "names_simple",
			typeDef: &ast.TypeDef{
				Name: "Status",
				Enum: &ast.Enum{Names: []string{"Active", "Inactive"}},
			},
			expected: []string{"statusactive", "statusinactive"},
		},
		{
			name: "names_takes_precedence_over_values",
			typeDef: &ast.TypeDef{
				Name: "Status",
				Enum: &ast.Enum{Names: []string{"Custom"}, Values: []string{"other"}},
			},
			expected: []string{"statuscustom"},
		},
		{
			name: "names_special_chars_sanitized",
			typeDef: &ast.TypeDef{
				Name: "Status",
				Enum: &ast.Enum{Names: []string{"in-progress"}},
			},
			expected: []string{"statusinprogress"},
		},
		// Type name variations
		{
			name: "type_name_snake_case",
			typeDef: &ast.TypeDef{
				Name: "order_status",
				Enum: &ast.Enum{Values: []string{"pending"}},
			},
			expected: []string{"orderstatuspending"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := n.EnumNames(tt.typeDef)
			assert.Equal(t, tt.expected, got)
		})
	}
}

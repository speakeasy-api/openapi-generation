package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFields_MustAddField_Success(t *testing.T) {
	type args struct {
		fields []*FieldDef
	}
	tests := []struct {
		name string
		args args
		want Fields
	}{
		{
			name: "successfully sorts ascending on insert",
			args: args{
				fields: []*FieldDef{
					{
						Name: "b",
					},
					{
						Name: "a",
					},
					{
						Name: "d",
					},
					{
						Name: "c",
					},
				},
			},
			want: Fields{
				{
					Name: "a",
				},
				{
					Name: "b",
				},
				{
					Name: "c",
				},
				{
					Name: "d",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ff := Fields{}

			for _, f := range tt.args.fields {
				ff = ff.MustAddField(f, false)
			}

			assert.Equal(t, tt.want, ff)
		})
	}
}

func TestFieldDef_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fd   *FieldDef
		test func(t *testing.T, original, cloned *FieldDef)
	}{
		{
			name: "nil FieldDef",
			fd:   nil,
			test: func(t *testing.T, original, cloned *FieldDef) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name: "simple FieldDef",
			fd: &FieldDef{
				Name:         "fieldName",
				OriginalName: "originalFieldName",
				Nullable:     true,
				Optional:     false,
				ErrorMessage: true,
				Type: &TypeDef{
					Type: DataTypeString,
				},
			},
			test: func(t *testing.T, original, cloned *FieldDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.Name, cloned.Name)
				assert.Equal(t, original.OriginalName, cloned.OriginalName)
				assert.Equal(t, original.Nullable, cloned.Nullable)
				assert.Equal(t, original.Optional, cloned.Optional)
				assert.Equal(t, original.ErrorMessage, cloned.ErrorMessage)
				assert.NotSame(t, original.Type, cloned.Type)
				assert.Equal(t, original.Type.Type, cloned.Type.Type)
			},
		},
		{
			name: "FieldDef with Comments",
			fd: &FieldDef{
				Name: "commentedField",
				Comments: &Comment{
					Summary:     "Field summary",
					Description: "Field description",
				},
				Type: &TypeDef{
					Type: DataTypeString,
				},
			},
			test: func(t *testing.T, original, cloned *FieldDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.NotNil(t, cloned.Comments)
				assert.NotSame(t, original.Comments, cloned.Comments)
				assert.Equal(t, original.Comments.Summary, cloned.Comments.Summary)
				assert.Equal(t, original.Comments.Description, cloned.Comments.Description)
			},
		},
		{
			name: "FieldDef with Const and Default",
			fd: &FieldDef{
				Name: "constField",
				Const: &AnyValue{
					Value: "constantValue",
				},
				Default: &AnyValue{
					Value: "defaultValue",
				},
				Type: &TypeDef{
					Type: DataTypeString,
				},
			},
			test: func(t *testing.T, original, cloned *FieldDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.NotNil(t, cloned.Const)
				assert.NotSame(t, original.Const, cloned.Const)
				assert.Equal(t, original.Const.Value, cloned.Const.Value)
				assert.NotNil(t, cloned.Default)
				assert.NotSame(t, original.Default, cloned.Default)
				assert.Equal(t, original.Default.Value, cloned.Default.Value)
			},
		},
		{
			name: "FieldDef with ParameterIndex",
			fd: &FieldDef{
				Name:           "paramField",
				ParameterIndex: ptr(3),
				Type: &TypeDef{
					Type: DataTypeString,
				},
			},
			test: func(t *testing.T, original, cloned *FieldDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.NotNil(t, cloned.ParameterIndex)
				assert.NotSame(t, original.ParameterIndex, cloned.ParameterIndex)
				assert.Equal(t, *original.ParameterIndex, *cloned.ParameterIndex)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.fd.Clone()
			tt.test(t, tt.fd, cloned)
		})
	}
}

func TestFieldDef_IsTerraformEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		f1       *FieldDef
		f2       *FieldDef
		expected bool
	}{
		{
			name:     "nil FieldDefs are equal",
			f1:       nil,
			f2:       nil,
			expected: true,
		},
		{
			name:     "nil and non-nil are not equal",
			f1:       nil,
			f2:       &FieldDef{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
			expected: false,
		},
		{
			name:     "different nullable not equal",
			f1:       &FieldDef{Name: "field1", Nullable: true, Type: &TypeDef{Type: DataTypeString}},
			f2:       &FieldDef{Name: "field1", Nullable: false, Type: &TypeDef{Type: DataTypeString}},
			expected: false,
		},
		{
			name:     "different optional not equal",
			f1:       &FieldDef{Name: "field1", Optional: true, Type: &TypeDef{Type: DataTypeString}},
			f2:       &FieldDef{Name: "field1", Optional: false, Type: &TypeDef{Type: DataTypeString}},
			expected: false,
		},
		{
			name:     "different types not equal",
			f1:       &FieldDef{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
			f2:       &FieldDef{Name: "field1", Type: &TypeDef{Type: DataTypeInteger}},
			expected: false,
		},
		{
			name:     "equal fields with same properties",
			f1:       &FieldDef{Name: "field1", Nullable: true, Optional: false, Type: &TypeDef{Type: DataTypeString}},
			f2:       &FieldDef{Name: "field1", Nullable: true, Optional: false, Type: &TypeDef{Type: DataTypeString}},
			expected: true,
		},
		{
			name: "equal fields with nested types",
			f1: &FieldDef{
				Name: "field1",
				Type: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "nested", Type: &TypeDef{Type: DataTypeInteger}},
					},
				},
			},
			f2: &FieldDef{
				Name: "field1",
				Type: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "nested", Type: &TypeDef{Type: DataTypeInteger}},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.f1.IsTerraformEqual(tt.f2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFields_IsTerraformEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		f1       Fields
		f2       Fields
		expected bool
	}{
		{
			name:     "empty fields are equal",
			f1:       Fields{},
			f2:       Fields{},
			expected: true,
		},
		{
			name: "different length not equal",
			f1: Fields{
				{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
			},
			f2: Fields{
				{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
			},
			expected: false,
		},
		{
			name: "equal fields with same properties",
			f1: Fields{
				{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
			},
			f2: Fields{
				{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
			},
			expected: true,
		},
		{
			name: "different field names not equal",
			f1: Fields{
				{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
			},
			f2: Fields{
				{Name: "field2", Type: &TypeDef{Type: DataTypeString}},
			},
			expected: false,
		},
		{
			name: "different field order is equal (compares by name)",
			f1: Fields{
				{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
			},
			f2: Fields{
				{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
				{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
			},
			expected: true,
		},
		{
			name: "equal complex fields",
			f1: Fields{
				{
					Name:     "field1",
					Nullable: true,
					Optional: false,
					Type:     &TypeDef{Type: DataTypeString},
				},
				{
					Name:     "field2",
					Nullable: false,
					Optional: true,
					Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "nested", Type: &TypeDef{Type: DataTypeInteger}},
						},
					},
				},
			},
			f2: Fields{
				{
					Name:     "field1",
					Nullable: true,
					Optional: false,
					Type:     &TypeDef{Type: DataTypeString},
				},
				{
					Name:     "field2",
					Nullable: false,
					Optional: true,
					Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "nested", Type: &TypeDef{Type: DataTypeInteger}},
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.f1.IsTerraformEqual(tt.f2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

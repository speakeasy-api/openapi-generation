package ast

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/stretchr/testify/assert"
)

func TestFieldDef_FindTerraformEquivalentField(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fieldDef       *FieldDef
		fields         Fields
		useMatchConfig bool
		expectedField  *FieldDef
		expectedPath   []string
	}{
		"nil receiver": {
			fieldDef: nil,
			fields: Fields{
				{Name: "test", Type: &TypeDef{Type: DataTypeString}},
			},
			useMatchConfig: false,
			expectedField:  nil,
			expectedPath:   nil,
		},
		"nil fields": {
			fieldDef: &FieldDef{
				Name: "test",
				Type: &TypeDef{Type: DataTypeString},
			},
			fields:         nil,
			useMatchConfig: false,
			expectedField:  nil,
			expectedPath:   nil,
		},
		"empty fields": {
			fieldDef: &FieldDef{
				Name: "test",
				Type: &TypeDef{Type: DataTypeString},
			},
			fields:         Fields{},
			useMatchConfig: false,
			expectedField:  nil,
			expectedPath:   nil,
		},
		"exact name match": {
			fieldDef: &FieldDef{
				Name: "test_field",
				Type: &TypeDef{Type: DataTypeString},
			},
			fields: Fields{
				{Name: "test_field", Type: &TypeDef{Type: DataTypeString}},
			},
			useMatchConfig: false,
			expectedPath:   []string{SanitizeFieldName("test_field")},
		},
		"sanitized name match": {
			fieldDef: &FieldDef{
				Name: "test_field",
				Type: &TypeDef{Type: DataTypeString},
			},
			fields: Fields{
				{Name: "TestField", Type: &TypeDef{Type: DataTypeString}},
			},
			useMatchConfig: false,
			expectedPath:   []string{SanitizeFieldName("TestField")},
		},
		"no match": {
			fieldDef: &FieldDef{
				Name: "test_field",
				Type: &TypeDef{Type: DataTypeString},
			},
			fields: Fields{
				{Name: "other_field", Type: &TypeDef{Type: DataTypeString}},
			},
			useMatchConfig: false,
			expectedField:  nil,
			expectedPath:   nil,
		},
		"match config disabled ignores match path": {
			fieldDef: &FieldDef{
				Name: "identifier",
				Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						MatchConfig: &extensions.MatchConfig{
							Path: ptr("nested.id"),
						},
					},
				},
			},
			fields: Fields{
				{
					Name: "nested",
					Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "id", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			useMatchConfig: false,
			expectedField:  nil,
			expectedPath:   nil,
		},
		"match config resolves single-step path": {
			fieldDef: &FieldDef{
				Name: "identifier",
				Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						MatchConfig: &extensions.MatchConfig{
							Path: ptr("resource_id"),
						},
					},
				},
			},
			fields: Fields{
				{Name: "resource_id", Type: &TypeDef{Type: DataTypeString}},
			},
			useMatchConfig: true,
			expectedPath:   []string{SanitizeFieldName("resource_id")},
		},
		"match config resolves multi-step path": {
			fieldDef: &FieldDef{
				Name: "identifier",
				Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						MatchConfig: &extensions.MatchConfig{
							Path: ptr("data.nested_id"),
						},
					},
				},
			},
			fields: Fields{
				{
					Name: "data",
					Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "nested_id", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			useMatchConfig: true,
			expectedPath:   []string{SanitizeFieldName("data"), SanitizeFieldName("nested_id")},
		},
		"match config path not found falls back to name match": {
			fieldDef: &FieldDef{
				Name: "test_field",
				Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						MatchConfig: &extensions.MatchConfig{
							Path: ptr("nonexistent.field"),
						},
					},
				},
			},
			fields: Fields{
				{Name: "test_field", Type: &TypeDef{Type: DataTypeString}},
			},
			useMatchConfig: true,
			expectedPath:   []string{SanitizeFieldName("test_field")},
		},
		"match config path not found and no name match": {
			fieldDef: &FieldDef{
				Name: "identifier",
				Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						MatchConfig: &extensions.MatchConfig{
							Path: ptr("nonexistent.field"),
						},
					},
				},
			},
			fields: Fields{
				{Name: "other", Type: &TypeDef{Type: DataTypeString}},
			},
			useMatchConfig: true,
			expectedField:  nil,
			expectedPath:   nil,
		},
		"match config with nil Path falls back to name match": {
			fieldDef: &FieldDef{
				Name: "test_field",
				Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						MatchConfig: &extensions.MatchConfig{
							UsePriorState: true,
						},
					},
				},
			},
			fields: Fields{
				{Name: "test_field", Type: &TypeDef{Type: DataTypeString}},
			},
			useMatchConfig: true,
			expectedPath:   []string{SanitizeFieldName("test_field")},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotField, gotPath := tc.fieldDef.FindTerraformEquivalentField(tc.fields, tc.useMatchConfig)

			switch {
			case tc.expectedField != nil:
				assert.Equal(t, tc.expectedField, gotField)
			case tc.expectedPath != nil:
				// When expectedField is not explicitly set but expectedPath is,
				// we just verify a field was returned.
				assert.NotNil(t, gotField)
			default:
				assert.Nil(t, gotField)
			}

			assert.Equal(t, tc.expectedPath, gotPath)
		})
	}
}

func TestFieldDef_IsTerraformImportRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fieldDef *FieldDef
		expected bool
	}{
		"nil receiver": {
			fieldDef: nil,
			expected: false,
		},
		"required - not optional not nullable": {
			fieldDef: &FieldDef{
				Name: "test",
				Type: &TypeDef{Type: DataTypeString},
			},
			expected: true,
		},
		"not required - optional": {
			fieldDef: &FieldDef{
				Name:     "test",
				Type:     &TypeDef{Type: DataTypeString},
				Optional: true,
			},
			expected: false,
		},
		"not required - nullable": {
			fieldDef: &FieldDef{
				Name:     "test",
				Type:     &TypeDef{Type: DataTypeString},
				Nullable: true,
			},
			expected: false,
		},
		"not required - optional and nullable": {
			fieldDef: &FieldDef{
				Name:     "test",
				Type:     &TypeDef{Type: DataTypeString},
				Optional: true,
				Nullable: true,
			},
			expected: false,
		},
		"required - annotation overrides optional": {
			fieldDef: &FieldDef{
				Name:     "test",
				Type:     &TypeDef{Type: DataTypeString},
				Optional: true,
				Annotations: Annotations{
					&ParamAnnotation{
						RequiredForOperation: true,
					},
				},
			},
			expected: true,
		},
		"required - annotation overrides nullable": {
			fieldDef: &FieldDef{
				Name:     "test",
				Type:     &TypeDef{Type: DataTypeString},
				Nullable: true,
				Annotations: Annotations{
					&ParamAnnotation{
						RequiredForOperation: true,
					},
				},
			},
			expected: true,
		},
		"annotation without RequiredForOperation does not override": {
			fieldDef: &FieldDef{
				Name:     "test",
				Type:     &TypeDef{Type: DataTypeString},
				Optional: true,
				Annotations: Annotations{
					&ParamAnnotation{
						RequiredForOperation: false,
					},
				},
			},
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tc.fieldDef.IsTerraformImportRequired()

			assert.Equal(t, tc.expected, got)
		})
	}
}

func Test_findTerraformMatchPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fields           Fields
		matchPath        string
		expectedResult   *FieldDef
		expectedAccessor []string
	}{
		"single step": {
			fields: Fields{
				{Name: "my_field", Type: &TypeDef{Type: DataTypeString}},
			},
			matchPath:        "my_field",
			expectedAccessor: []string{SanitizeFieldName("my_field")},
		},
		"multi step": {
			fields: Fields{
				{
					Name: "root",
					Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{
								Name: "mid",
								Type: &TypeDef{
									Type: DataTypeClass,
									Fields: Fields{
										{Name: "leaf_field", Type: &TypeDef{Type: DataTypeString}},
									},
								},
							},
						},
					},
				},
			},
			matchPath: "root.mid.leaf_field",
			expectedAccessor: []string{
				SanitizeFieldName("root"),
				SanitizeFieldName("mid"),
				SanitizeFieldName("leaf_field"),
			},
		},
		"step not found": {
			fields: Fields{
				{Name: "existing", Type: &TypeDef{Type: DataTypeString}},
			},
			matchPath:        "nonexistent",
			expectedResult:   nil,
			expectedAccessor: nil,
		},
		"intermediate step not found": {
			fields: Fields{
				{
					Name: "root",
					Type: &TypeDef{
						Type:   DataTypeClass,
						Fields: Fields{},
					},
				},
			},
			matchPath:        "root.nonexistent",
			expectedResult:   nil,
			expectedAccessor: nil,
		},
		"sanitized name comparison": {
			fields: Fields{
				{Name: "MyField", Type: &TypeDef{Type: DataTypeString}},
			},
			matchPath:        "my_field",
			expectedAccessor: []string{SanitizeFieldName("MyField")},
		},
		"nil fields": {
			fields:           nil,
			matchPath:        "test",
			expectedResult:   nil,
			expectedAccessor: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotResult, gotAccessor := findTerraformMatchPath(tc.fields, tc.matchPath)

			switch {
			case tc.expectedResult != nil:
				assert.Equal(t, tc.expectedResult, gotResult)
			case tc.expectedAccessor != nil:
				assert.NotNil(t, gotResult)
			default:
				assert.Nil(t, gotResult)
			}

			assert.Equal(t, tc.expectedAccessor, gotAccessor)
		})
	}
}

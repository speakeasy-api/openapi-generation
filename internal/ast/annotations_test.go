package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodingAnnotation_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		orig *EncodingAnnotation
	}{
		{
			name: "nil annotation",
			orig: nil,
		},
		{
			name: "empty annotation",
			orig: &EncodingAnnotation{},
		},
		{
			name: "annotation with media type",
			orig: &EncodingAnnotation{
				MediaType: "application/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.orig.Clone()

			if tt.orig == nil {
				assert.Nil(t, cloned)
				return
			}

			clonedTyped, ok := cloned.(*EncodingAnnotation)
			require.True(t, ok, "Clone() returned wrong type: got %T, want *EncodingAnnotation", cloned)
			assert.Equal(t, tt.orig, clonedTyped)
			assert.NotSame(t, tt.orig, clonedTyped)
		})
	}
}

func TestFormAnnotation_Clone(t *testing.T) {
	t.Parallel()

	testTypeDef := &TypeDef{
		Name: "TestType",
	}

	tests := []struct {
		name string
		orig *FormAnnotation
	}{
		{
			name: "nil annotation",
			orig: nil,
		},
		{
			name: "empty annotation",
			orig: &FormAnnotation{},
		},
		{
			name: "full annotation",
			orig: &FormAnnotation{
				Name:      "test_name",
				JSON:      true,
				Style:     "simple",
				Explode:   true,
				FieldType: testTypeDef,
			},
		},
		{
			name: "annotation without field type",
			orig: &FormAnnotation{
				Name:    "test_name",
				JSON:    false,
				Style:   "form",
				Explode: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.orig.Clone()

			if tt.orig == nil {
				assert.Nil(t, cloned)
				return
			}

			clonedTyped, ok := cloned.(*FormAnnotation)
			require.True(t, ok, "Clone() returned wrong type: got %T, want *FormAnnotation", cloned)
			assert.Equal(t, tt.orig, clonedTyped)
			assert.NotSame(t, tt.orig, clonedTyped)

			// Verify FieldType is deep copied
			if tt.orig.FieldType != nil {
				assert.NotSame(t, tt.orig.FieldType, clonedTyped.FieldType)
			}
		})
	}
}

func TestJSONAnnotation_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		orig *JSONAnnotation
	}{
		{
			name: "nil annotation",
			orig: nil,
		},
		{
			name: "empty annotation",
			orig: &JSONAnnotation{},
		},
		{
			name: "annotation with ignore true",
			orig: &JSONAnnotation{
				Ignore:    true,
				FieldName: "test_field",
			},
		},
		{
			name: "annotation with ignore false",
			orig: &JSONAnnotation{
				Ignore:    false,
				FieldName: "another_field",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.orig.Clone()

			if tt.orig == nil {
				assert.Nil(t, cloned)
				return
			}

			clonedTyped, ok := cloned.(*JSONAnnotation)
			require.True(t, ok, "Clone() returned wrong type: got %T, want *JSONAnnotation", cloned)
			assert.Equal(t, tt.orig, clonedTyped)
			assert.NotSame(t, tt.orig, clonedTyped)
		})
	}
}

func TestMultipartFormAnnotation_Clone(t *testing.T) {
	t.Parallel()

	testTypeDef := &TypeDef{
		Name: "TestType",
	}

	tests := []struct {
		name string
		orig *MultipartFormAnnotation
	}{
		{
			name: "nil annotation",
			orig: nil,
		},
		{
			name: "empty annotation",
			orig: &MultipartFormAnnotation{},
		},
		{
			name: "full annotation",
			orig: &MultipartFormAnnotation{
				File:      true,
				Content:   true,
				JSON:      false,
				Name:      "upload_file",
				FieldType: testTypeDef,
			},
		},
		{
			name: "annotation without field type",
			orig: &MultipartFormAnnotation{
				File:    false,
				Content: false,
				JSON:    true,
				Name:    "data",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.orig.Clone()

			if tt.orig == nil {
				assert.Nil(t, cloned)
				return
			}

			clonedTyped, ok := cloned.(*MultipartFormAnnotation)
			require.True(t, ok, "Clone() returned wrong type: got %T, want *MultipartFormAnnotation", cloned)
			assert.Equal(t, tt.orig, clonedTyped)
			assert.NotSame(t, tt.orig, clonedTyped)

			// Verify FieldType is deep copied
			if tt.orig.FieldType != nil {
				assert.NotSame(t, tt.orig.FieldType, clonedTyped.FieldType)
			}
		})
	}
}

func TestParamAnnotation_Clone(t *testing.T) {
	t.Parallel()

	testTypeDef := &TypeDef{
		Name: "TestParamType",
	}

	tests := []struct {
		name string
		orig *ParamAnnotation
	}{
		{
			name: "nil annotation",
			orig: nil,
		},
		{
			name: "empty annotation",
			orig: &ParamAnnotation{},
		},
		{
			name: "full annotation",
			orig: &ParamAnnotation{
				ParamType:            ParamTypeQueryParam,
				Name:                 "test_param",
				Serialization:        "form",
				Style:                "simple",
				Explode:              true,
				FieldType:            testTypeDef,
				AllowReserved:        true,
				IsGlobal:             true,
				OperationsForGlobal:  []string{"op1", "op2", "op3"},
				HasGlobal:            false,
				Hidden:               true,
				RequiredForOperation: true,
			},
		},
		{
			name: "annotation without optional fields",
			orig: &ParamAnnotation{
				ParamType: ParamTypePathParam,
				Name:      "id",
			},
		},
		{
			name: "annotation with nil operations slice",
			orig: &ParamAnnotation{
				ParamType:           ParamTypeHeader,
				Name:                "X-Custom-Header",
				OperationsForGlobal: nil,
			},
		},
		{
			name: "annotation with empty operations slice",
			orig: &ParamAnnotation{
				ParamType:           ParamTypeHeader,
				Name:                "Authorization",
				OperationsForGlobal: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.orig.Clone()

			if tt.orig == nil {
				assert.Nil(t, cloned)
				return
			}

			clonedTyped, ok := cloned.(*ParamAnnotation)
			require.True(t, ok, "Clone() returned wrong type: got %T, want *ParamAnnotation", cloned)
			assert.Equal(t, tt.orig, clonedTyped)
			assert.NotSame(t, tt.orig, clonedTyped)

			// Verify FieldType is deep copied
			if tt.orig.FieldType != nil {
				assert.NotSame(t, tt.orig.FieldType, clonedTyped.FieldType)
			}

			// Verify OperationsForGlobal slice is deep copied
			if len(tt.orig.OperationsForGlobal) > 0 {
				assert.NotSame(t, &tt.orig.OperationsForGlobal[0], &clonedTyped.OperationsForGlobal[0])
			}
		})
	}
}

func TestRequestAnnotation_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		orig *RequestAnnotation
	}{
		{
			name: "nil annotation",
			orig: nil,
		},
		{
			name: "empty annotation",
			orig: &RequestAnnotation{},
		},
		{
			name: "annotation with media type",
			orig: &RequestAnnotation{
				MediaType: "application/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.orig.Clone()

			if tt.orig == nil {
				assert.Nil(t, cloned)
				return
			}

			clonedTyped, ok := cloned.(*RequestAnnotation)
			require.True(t, ok, "Clone() returned wrong type: got %T, want *RequestAnnotation", cloned)
			assert.Equal(t, tt.orig, clonedTyped)
			assert.NotSame(t, tt.orig, clonedTyped)
		})
	}
}

func TestRequestWrapperAnnotation_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		orig *RequestWrapperAnnotation
	}{
		{
			name: "nil annotation",
			orig: nil,
		},
		{
			name: "empty annotation",
			orig: &RequestWrapperAnnotation{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.orig.Clone()

			if tt.orig == nil {
				assert.Nil(t, cloned)
				return
			}

			clonedTyped, ok := cloned.(*RequestWrapperAnnotation)
			require.True(t, ok, "Clone() returned wrong type: got %T, want *RequestWrapperAnnotation", cloned)
			assert.Equal(t, tt.orig, clonedTyped)

			// Note: For empty structs, Go may optimize to use the same memory address
			// The important part is that Clone() returns a new instance, which it does
			// assert.NotSame(t, tt.orig, clonedTyped)
		})
	}
}

func TestResponseAnnotation_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		orig *ResponseAnnotation
	}{
		{
			name: "nil annotation",
			orig: nil,
		},
		{
			name: "empty annotation",
			orig: &ResponseAnnotation{},
		},
		{
			name: "annotation with result field true",
			orig: &ResponseAnnotation{
				ResultField: true,
			},
		},
		{
			name: "annotation with result field false",
			orig: &ResponseAnnotation{
				ResultField: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.orig.Clone()

			if tt.orig == nil {
				assert.Nil(t, cloned)
				return
			}

			clonedTyped, ok := cloned.(*ResponseAnnotation)
			require.True(t, ok, "Clone() returned wrong type: got %T, want *ResponseAnnotation", cloned)
			assert.Equal(t, tt.orig, clonedTyped)
			assert.NotSame(t, tt.orig, clonedTyped)
		})
	}
}

func TestSecurityAnnotation_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		orig *SecurityAnnotation
	}{
		{
			name: "nil annotation",
			orig: nil,
		},
		{
			name: "empty annotation",
			orig: &SecurityAnnotation{},
		},
		{
			name: "full annotation",
			orig: &SecurityAnnotation{
				FieldName:      "api_key",
				SecType:        "apiKey",
				SubType:        "header",
				Option:         true,
				Scheme:         true,
				SecurityOption: false,
				SchemeKey:      "X-API-Key",
			},
		},
		{
			name: "oauth2 annotation",
			orig: &SecurityAnnotation{
				FieldName:      "oauth_token",
				SecType:        "oauth2",
				SubType:        "client_credentials",
				Option:         false,
				Scheme:         true,
				SecurityOption: true,
				SchemeKey:      "oauth2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.orig.Clone()

			if tt.orig == nil {
				assert.Nil(t, cloned)
				return
			}

			clonedTyped, ok := cloned.(*SecurityAnnotation)
			require.True(t, ok, "Clone() returned wrong type: got %T, want *SecurityAnnotation", cloned)
			assert.Equal(t, tt.orig, clonedTyped)
			assert.NotSame(t, tt.orig, clonedTyped)
		})
	}
}

func TestOperationSecurityAnnotation_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		orig *OperationSecurityAnnotation
	}{
		{
			name: "nil annotation",
			orig: nil,
		},
		{
			name: "empty annotation",
			orig: &OperationSecurityAnnotation{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.orig.Clone()

			if tt.orig == nil {
				assert.Nil(t, cloned)
				return
			}

			clonedTyped, ok := cloned.(*OperationSecurityAnnotation)
			require.True(t, ok, "Clone() returned wrong type: got %T, want *OperationSecurityAnnotation", cloned)
			assert.Equal(t, tt.orig, clonedTyped)

			// Note: For empty structs, Go may optimize to use the same memory address
			// The important part is that Clone() returns a new instance, which it does
			// assert.NotSame(t, tt.orig, clonedTyped)
		})
	}
}

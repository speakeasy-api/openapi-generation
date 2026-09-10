package schemas

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/stretchr/testify/assert"
)

func TestDefaultValueMatchesType(t *testing.T) {
	tests := []struct {
		name     string
		val      any
		dataType ast.DataType
		want     bool
	}{
		{"int matches integer", int(42), ast.DataTypeInteger, true},
		{"int matches number", int(0), ast.DataTypeNumber, true},
		{"float64 matches number", float64(3.14), ast.DataTypeNumber, true},
		{"float64 does not match integer", float64(3.14), ast.DataTypeInteger, false},
		{"string matches string", "hello", ast.DataTypeString, true},
		{"string does not match number", "0.00", ast.DataTypeNumber, false},
		{"bool matches boolean", true, ast.DataTypeBoolean, true},
		{"string does not match boolean", "true", ast.DataTypeBoolean, false},
		{"int does not match string", int(0), ast.DataTypeString, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultValueMatchesType(tt.val, tt.dataType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCoerceDefaultValue(t *testing.T) {
	tests := []struct {
		name     string
		val      any
		dataType ast.DataType
		wantVal  any
		wantOK   bool
	}{
		// String to number coercions
		{"string '0.00' to number", "0.00", ast.DataTypeNumber, float64(0), true},
		{"string '1.5' to number", "1.5", ast.DataTypeNumber, float64(1.5), true},
		{"string '123' to number", "123", ast.DataTypeNumber, float64(123), true},
		{"string '0.00' to float32", "0.00", ast.DataTypeFloat32, float64(0), true},
		{"string '0.00' to decimal", "0.00", ast.DataTypeDecimal, float64(0), true},

		// String to integer coercions
		{"string '42' to integer", "42", ast.DataTypeInteger, int(42), true},
		{"string '0' to integer", "0", ast.DataTypeInteger, int(0), true},
		{"string '1.0' to integer (whole number)", "1.0", ast.DataTypeInteger, int(1), true},
		{"string '1.5' to integer (not whole)", "1.5", ast.DataTypeInteger, nil, false},
		{"string '42' to int32", "42", ast.DataTypeInt32, int(42), true},

		// String to boolean coercions
		{"string 'true' to boolean", "true", ast.DataTypeBoolean, true, true},
		{"string 'false' to boolean", "false", ast.DataTypeBoolean, false, true},
		{"string 'True' to boolean", "True", ast.DataTypeBoolean, true, true},
		{"string 'FALSE' to boolean", "FALSE", ast.DataTypeBoolean, false, true},
		{"string 'variant' to boolean", "variant", ast.DataTypeBoolean, nil, false},

		// Float to integer coercions
		{"float64(0) to integer", float64(0), ast.DataTypeInteger, int(0), true},
		{"float64(42) to integer", float64(42), ast.DataTypeInteger, int(42), true},
		{"float64(1.5) to integer (not whole)", float64(1.5), ast.DataTypeInteger, nil, false},
		{"float32(0) to integer", float32(0), ast.DataTypeInteger, int(0), true},

		// Non-coercible cases
		{"string 'hello' to number", "hello", ast.DataTypeNumber, nil, false},
		{"string '' to number", "", ast.DataTypeNumber, nil, false},
		{"bool to number (no coercion)", true, ast.DataTypeNumber, nil, false},
		{"string to string (not needed)", "hello", ast.DataTypeString, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOK := coerceDefaultValue(tt.val, tt.dataType)
			assert.Equal(t, tt.wantOK, gotOK)
			if tt.wantOK {
				assert.Equal(t, tt.wantVal, gotVal)
			}
		})
	}
}

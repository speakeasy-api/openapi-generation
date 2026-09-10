package terraform

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// JSONToHCLExpression converts a JSON string into an HCL jsonencode()
// expression. For example, the JSON string `{"key": "value"}` becomes:
//
//	jsonencode({
//	  key = "value"
//	})
//
// If conversion fails for any reason, the original JSON string is returned
// as a quoted HCL string literal.
func JSONToHCLExpression(jsonStr string) string {
	var raw any

	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return fmt.Sprintf("%q", jsonStr)
	}

	val, err := jsonValueToCty(raw)
	if err != nil {
		return fmt.Sprintf("%q", jsonStr)
	}

	valueTokens := hclwrite.TokensForValue(val)
	callTokens := hclwrite.TokensForFunctionCall("jsonencode", valueTokens)

	return string(callTokens.Bytes())
}

// jsonValueToCty recursively converts a decoded JSON value (from
// encoding/json) into a cty.Value suitable for hclwrite.TokensForValue.
func jsonValueToCty(v any) (cty.Value, error) {
	switch val := v.(type) {
	case nil:
		return cty.NullVal(cty.DynamicPseudoType), nil
	case bool:
		return cty.BoolVal(val), nil
	case float64:
		return cty.NumberVal(new(big.Float).SetFloat64(val)), nil
	case string:
		return cty.StringVal(val), nil
	case []any:
		if len(val) == 0 {
			return cty.EmptyTupleVal, nil
		}

		elems := make([]cty.Value, len(val))

		for i, item := range val {
			elem, err := jsonValueToCty(item)
			if err != nil {
				return cty.NilVal, err
			}

			elems[i] = elem
		}

		return cty.TupleVal(elems), nil
	case map[string]any:
		if len(val) == 0 {
			return cty.EmptyObjectVal, nil
		}

		attrs := make(map[string]cty.Value, len(val))

		for k, item := range val {
			attr, err := jsonValueToCty(item)
			if err != nil {
				return cty.NilVal, err
			}

			attrs[k] = attr
		}

		return cty.ObjectVal(attrs), nil
	default:
		return cty.NilVal, fmt.Errorf("unsupported JSON type: %T", v)
	}
}

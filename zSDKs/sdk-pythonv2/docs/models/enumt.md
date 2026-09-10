# EnumT

## Example Usage

```python
from speakeasy.new_openapi.models import EnumT

# Open enum: unrecognized values are captured as UnrecognizedStr
value: EnumT = "First"
```


## Values

This is an open enum. Unrecognized values will not fail type checks.

- `"First"`
- `"Second"`
- `"Ex-aequo"`
- `"Ex.aequo"`

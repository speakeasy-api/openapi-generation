# IntEnum

An integer enum property.

## Example Usage

```python
from speakeasy.new_openapi.models import IntEnum

# Open enum: unrecognized values are captured as UnrecognizedInt
value: IntEnum = 1
```


## Values

This is an open enum. Unrecognized values will not fail type checks.

- `1`
- `2`
- `3`

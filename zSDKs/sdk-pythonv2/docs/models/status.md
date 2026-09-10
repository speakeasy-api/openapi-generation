# Status

## Example Usage

```python
from speakeasy.new_openapi.models import Status

# Open enum: unrecognized values are captured as UnrecognizedStr
value: Status = "in_progress"
```


## Values

This is an open enum. Unrecognized values will not fail type checks.

- `"in_progress"`
- `"completed"`
- `"failed"`
- `"requires_action"`

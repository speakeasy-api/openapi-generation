# Status

## Example Usage

```go
import (
	"example.com/openapi-go-sdk"
)

value := examplealias.StatusInProgress

// Open enum: custom values can be created with a direct type cast
custom := examplealias.Status("custom_value")
```


## Values

| Name                   | Value                  |
| ---------------------- | ---------------------- |
| `StatusInProgress`     | in_progress            |
| `StatusCompleted`      | completed              |
| `StatusFailed`         | failed                 |
| `StatusRequiresAction` | requires_action        |
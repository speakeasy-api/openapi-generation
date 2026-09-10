# PetStatus

Status of a pet in the bar namespace

## Example Usage

```go
import (
	"example.com/openapi-go-sdk/bar"
)

value := bar.PetStatusAvailable

// Open enum: custom values can be created with a direct type cast
custom := bar.PetStatus("custom_value")
```


## Values

| Name                   | Value                  |
| ---------------------- | ---------------------- |
| `PetStatusAvailable`   | available              |
| `PetStatusAdopted`     | adopted                |
| `PetStatusPending`     | pending                |
| `PetStatusUnavailable` | unavailable            |
# IntEnum

An integer enum property.

## Example Usage

```csharp
using Speakeasy.OpenAPI;

var value = IntEnum.First;

// Open enum: use .Of() to create instances from custom integer values
var custom = IntEnum.Of(999);
```


## Values

| Name     | Value    |
| -------- | -------- |
| `First`  | 1        |
| `Second` | 2        |
| `Third`  | 3        |
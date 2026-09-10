# Int32Enum

An int32 enum property.

## Example Usage

```csharp
using Speakeasy.OpenAPI;

var value = Int32Enum.FiftyFive;

// Open enum: use .Of() to create instances from custom integer values
var custom = Int32Enum.Of(999);
```


## Values

| Name                     | Value                    |
| ------------------------ | ------------------------ |
| `FiftyFive`              | 55                       |
| `SixtyNine`              | 69                       |
| `OneHundredAndEightyOne` | 181                      |
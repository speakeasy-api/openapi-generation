# Enum

## Example Usage

```csharp
using Speakeasy.OpenAPI;

var value = Enum.First;

// Open enum: use .Of() to create instances from custom string values
var custom = Enum.Of("custom_value");
```


## Values

| Name     | Value    |
| -------- | -------- |
| `First`  | First    |
| `Second` | Second   |
| `ThirdA` | Ex-aequo |
| `ThirdB` | Ex.aequo |
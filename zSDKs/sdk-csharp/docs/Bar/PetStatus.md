# PetStatus

Status of a pet in the bar namespace

## Example Usage

```csharp
using Speakeasy.OpenAPI.Bar;

var value = PetStatus.Available;

// Open enum: use .Of() to create instances from custom string values
var custom = PetStatus.Of("custom_value");
```


## Values

| Name          | Value         |
| ------------- | ------------- |
| `Available`   | available     |
| `Adopted`     | adopted       |
| `Pending`     | pending       |
| `Unavailable` | unavailable   |
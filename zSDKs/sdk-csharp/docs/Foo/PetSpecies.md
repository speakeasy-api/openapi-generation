# PetSpecies

Species of a pet in the foo namespace

## Example Usage

```csharp
using Speakeasy.OpenAPI.Foo;

var value = PetSpecies.Dog;

// Open enum: use .Of() to create instances from custom string values
var custom = PetSpecies.Of("custom_value");
```


## Values

| Name   | Value  |
| ------ | ------ |
| `Dog`  | dog    |
| `Cat`  | cat    |
| `Bird` | bird   |
| `Fish` | fish   |
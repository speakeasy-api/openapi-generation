# Animal

A discriminated union of animal types in the foo namespace


## Supported Types

### Dog

```go
animal := examplealias.NewAnimal(foo.Dog{/* values here */})
```

### Cat

```go
animal := examplealias.NewAnimal(foo.Cat{/* values here */})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch animal.Type {
	case examplealias.AnimalTypeDog:
		// animal.Dog is populated
	case examplealias.AnimalTypeCat:
		// animal.Cat is populated
	default:
		// Unknown type - use animal.GetUnknownRaw() for raw JSON
}
```

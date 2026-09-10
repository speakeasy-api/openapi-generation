# OneOfWithUnionDescription

A union of two types.


## Supported Types

### ExhaustiveObject

```go
oneOfWithUnionDescription := examplealias.NewOneOfWithUnionDescription(ExhaustiveObject{/* values here */})
```

### SimpleObject

```go
oneOfWithUnionDescription := examplealias.NewOneOfWithUnionDescription(SimpleObject{/* values here */})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch oneOfWithUnionDescription.Type {
	case examplealias.OneOfWithUnionDescriptionTypeExhaustiveObject:
		// oneOfWithUnionDescription.ExhaustiveObject is populated
	case examplealias.OneOfWithUnionDescriptionTypeSimpleObject:
		// oneOfWithUnionDescription.SimpleObject is populated
	default:
		// Unknown type - use oneOfWithUnionDescription.GetUnknownRaw() for raw JSON
}
```

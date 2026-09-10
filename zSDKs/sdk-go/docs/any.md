# Any


## Supported Types

### SimpleObject

```go
any := examplealias.NewAny(SimpleObject{/* values here */})
```

### 

```go
any := examplealias.NewAny("")
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch any.Type {
	case examplealias.AnyTypeSimpleObject:
		// any.SimpleObject is populated
	case examplealias.AnyTypeStr:
		// any.Str is populated
	default:
		// Unknown type - use any.GetUnknownRaw() for raw JSON
}
```

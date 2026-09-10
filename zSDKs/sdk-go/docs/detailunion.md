# DetailUnion

Contains parameter or domain specific information related to the error and why it occurred.


## Supported Types

### 

```go
detailUnion := examplealias.NewDetailUnion("")
```

### Detail

```go
detailUnion := examplealias.NewDetailUnion(Detail{/* values here */})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch detailUnion.Type {
	case examplealias.DetailUnionTypeStr:
		// detailUnion.Str is populated
	case examplealias.DetailUnionTypeDetail:
		// detailUnion.Detail is populated
	default:
		// Unknown type - use detailUnion.GetUnknownRaw() for raw JSON
}
```

# CircularUnion


## Supported Types

### 

```go
circularUnion := examplealias.NewCircularUnion(map[string]CircularUnion{/* values here */})
```

### 

```go
circularUnion := examplealias.NewCircularUnion("")
```

### 

```go
circularUnion := examplealias.NewCircularUnion(int64(0))
```

### 

```go
circularUnion := examplealias.NewCircularUnion(false)
```

### 

```go
circularUnion := examplealias.NewCircularUnion([]CircularUnion{/* values here */})
```

### 

```go
circularUnion := examplealias.NewCircularUnion(float64(0))
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch circularUnion.Type {
	case examplealias.CircularUnionTypeMapOfCircularUnion:
		// circularUnion.MapOfCircularUnion is populated
	case examplealias.CircularUnionTypeStr:
		// circularUnion.Str is populated
	case examplealias.CircularUnionTypeInteger:
		// circularUnion.Integer is populated
	case examplealias.CircularUnionTypeBoolean:
		// circularUnion.Boolean is populated
	case examplealias.CircularUnionTypeArrayOfCircularUnion:
		// circularUnion.ArrayOfCircularUnion is populated
	case examplealias.CircularUnionTypeNumber:
		// circularUnion.Number is populated
	default:
		// Unknown type - use circularUnion.GetUnknownRaw() for raw JSON
}
```

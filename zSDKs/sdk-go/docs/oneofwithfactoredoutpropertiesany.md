# OneOfWithFactoredOutPropertiesAny


## Supported Types

### SimpleObject

```go
oneOfWithFactoredOutPropertiesAny := examplealias.NewOneOfWithFactoredOutPropertiesAny(SimpleObject{/* values here */})
```

### 

```go
oneOfWithFactoredOutPropertiesAny := examplealias.NewOneOfWithFactoredOutPropertiesAny("")
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch oneOfWithFactoredOutPropertiesAny.Type {
	case examplealias.OneOfWithFactoredOutPropertiesAnyTypeSimpleObject:
		// oneOfWithFactoredOutPropertiesAny.SimpleObject is populated
	case examplealias.OneOfWithFactoredOutPropertiesAnyTypeStr:
		// oneOfWithFactoredOutPropertiesAny.Str is populated
	default:
		// Unknown type - use oneOfWithFactoredOutPropertiesAny.GetUnknownRaw() for raw JSON
}
```

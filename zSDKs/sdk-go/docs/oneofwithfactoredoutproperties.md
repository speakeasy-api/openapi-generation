# OneOfWithFactoredOutProperties

A union of two types with factored out properties.


## Supported Types

### OneOfWithFactoredOutPropertiesExhaustiveObject

```go
oneOfWithFactoredOutProperties := examplealias.NewOneOfWithFactoredOutProperties(OneOfWithFactoredOutPropertiesExhaustiveObject{/* values here */})
```

### OneOfWithFactoredOutPropertiesSimpleObject

```go
oneOfWithFactoredOutProperties := examplealias.NewOneOfWithFactoredOutProperties(OneOfWithFactoredOutPropertiesSimpleObject{/* values here */})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch oneOfWithFactoredOutProperties.Type {
	case examplealias.OneOfWithFactoredOutPropertiesTypeOneOfWithFactoredOutPropertiesExhaustiveObject:
		// oneOfWithFactoredOutProperties.OneOfWithFactoredOutPropertiesExhaustiveObject is populated
	case examplealias.OneOfWithFactoredOutPropertiesTypeOneOfWithFactoredOutPropertiesSimpleObject:
		// oneOfWithFactoredOutProperties.OneOfWithFactoredOutPropertiesSimpleObject is populated
	default:
		// Unknown type - use oneOfWithFactoredOutProperties.GetUnknownRaw() for raw JSON
}
```

# Shape

A discriminated union of shape types


## Supported Types

### Circle

```go
shape := examplealias.NewShape(Circle{/* values here */})
```

### Rectangle

```go
shape := examplealias.NewShape(Rectangle{/* values here */})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch shape.Type {
	case examplealias.ShapeTypeCircle:
		// shape.Circle is populated
	case examplealias.ShapeTypeRectangle:
		// shape.Rectangle is populated
	default:
		// Unknown type - use shape.GetUnknownRaw() for raw JSON
}
```

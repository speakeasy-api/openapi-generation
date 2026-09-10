# SomeUnion

A union demonstrating title and x-speakeasy-name-override on primitive types


## Supported Types

### MyString

```go
someUnion := examplealias.NewSomeUnion("")
```

### MyObject

```go
someUnion := examplealias.NewSomeUnion(MyObject{/* values here */})
```

### Foo

```go
someUnion := examplealias.NewSomeUnion(false)
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch someUnion.Type {
	case examplealias.SomeUnionTypeMyString:
		// someUnion.MyString is populated
	case examplealias.SomeUnionTypeMyObject:
		// someUnion.MyObject is populated
	case examplealias.SomeUnionTypeFoo:
		// someUnion.Foo is populated
	default:
		// Unknown type - use someUnion.GetUnknownRaw() for raw JSON
}
```

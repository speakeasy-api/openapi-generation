# ClientError

Something went wrong


## Supported Types

### TaggedError1

```go
clientError := examplealias.NewClientError(TaggedError1{})
```

### TaggedError2

```go
clientError := examplealias.NewClientError(TaggedError2{})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch clientError.Type {
	case examplealias.ClientErrorTypeTag1:
		// clientError.TaggedError1 is populated
	case examplealias.ClientErrorTypeTag2:
		// clientError.TaggedError2 is populated
	default:
		// Unknown type - use clientError.GetUnknownRaw() for raw JSON
}
```

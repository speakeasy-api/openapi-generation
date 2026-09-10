# GetErrorInUnionInternalServerError

OK


## Supported Types

### ErrorsError

```go
getErrorInUnionInternalServerError := examplealias.NewGetErrorInUnionInternalServerError(ErrorsError{})
```

### TaggedError1

```go
getErrorInUnionInternalServerError := examplealias.NewGetErrorInUnionInternalServerError(TaggedError1{})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch getErrorInUnionInternalServerError.Type {
	case examplealias.GetErrorInUnionInternalServerErrorTypeErrorsError:
		// getErrorInUnionInternalServerError.ErrorsError is populated
	case examplealias.GetErrorInUnionInternalServerErrorTypeTaggedError1:
		// getErrorInUnionInternalServerError.TaggedError1 is populated
	default:
		// Unknown type - use getErrorInUnionInternalServerError.GetUnknownRaw() for raw JSON
}
```

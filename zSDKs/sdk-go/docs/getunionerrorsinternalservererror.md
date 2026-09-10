# GetUnionErrorsInternalServerError

Internal Server Error


## Supported Types

### ErrorType1

```go
getUnionErrorsInternalServerError := examplealias.NewGetUnionErrorsInternalServerError(ErrorType1{})
```

### ErrorType2

```go
getUnionErrorsInternalServerError := examplealias.NewGetUnionErrorsInternalServerError(ErrorType2{})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch getUnionErrorsInternalServerError.Type {
	case examplealias.GetUnionErrorsInternalServerErrorTypeErrorType1:
		// getUnionErrorsInternalServerError.ErrorType1 is populated
	case examplealias.GetUnionErrorsInternalServerErrorTypeErrorType2:
		// getUnionErrorsInternalServerError.ErrorType2 is populated
	default:
		// Unknown type - use getUnionErrorsInternalServerError.GetUnknownRaw() for raw JSON
}
```

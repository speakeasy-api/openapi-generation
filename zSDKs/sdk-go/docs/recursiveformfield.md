# RecursiveFormField

A form field schema that can be recursive - the array type variant contains an array of RecursiveFormField items.


## Supported Types

### RecursiveFormFieldText

```go
recursiveFormField := examplealias.NewRecursiveFormField(RecursiveFormFieldText{/* values here */})
```

### RecursiveFormFieldNumber

```go
recursiveFormField := examplealias.NewRecursiveFormField(RecursiveFormFieldNumber{/* values here */})
```

### RecursiveFormFieldArray

```go
recursiveFormField := examplealias.NewRecursiveFormField(RecursiveFormFieldArray{/* values here */})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch recursiveFormField.Type {
	case examplealias.RecursiveFormFieldTypeText:
		// recursiveFormField.RecursiveFormFieldText is populated
	case examplealias.RecursiveFormFieldTypeNumber:
		// recursiveFormField.RecursiveFormFieldNumber is populated
	case examplealias.RecursiveFormFieldTypeArray:
		// recursiveFormField.RecursiveFormFieldArray is populated
	default:
		// Unknown type - use recursiveFormField.GetUnknownRaw() for raw JSON
}
```

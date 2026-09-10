# Content


## Supported Types

### AssetTextBlock

```go
content := examplealias.NewContent(AssetTextBlock{/* values here */})
```

### AssetImageBlock

```go
content := examplealias.NewContent(AssetImageBlock{/* values here */})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch content.Type {
	case examplealias.ContentTypeText:
		// content.AssetTextBlock is populated
	case examplealias.ContentTypeImage:
		// content.AssetImageBlock is populated
	default:
		// Unknown type - use content.GetUnknownRaw() for raw JSON
}
```

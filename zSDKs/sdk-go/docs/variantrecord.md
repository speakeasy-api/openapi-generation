# VariantRecord

A discriminated union with $-prefixed keys and x-speakeasy-discriminator overrides


## Supported Types

### TextRecord

```go
variantRecord := examplealias.NewVariantRecord(TextRecord{/* values here */})
```

### ImageRecord

```go
variantRecord := examplealias.NewVariantRecord(ImageRecord{/* values here */})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch variantRecord.Type {
	case examplealias.VariantRecordTypeText:
		// variantRecord.TextRecord is populated
	case examplealias.VariantRecordTypeImage:
		// variantRecord.ImageRecord is populated
}
```

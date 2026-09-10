# DisriminatedUnionWithOneMember


## Supported Types

### ExhaustiveObject

```go
disriminatedUnionWithOneMember := examplealias.NewDisriminatedUnionWithOneMember(ExhaustiveObject{/* values here */})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch disriminatedUnionWithOneMember.Type {
	case examplealias.DisriminatedUnionWithOneMemberTypeType1:
		// disriminatedUnionWithOneMember.ExhaustiveObject is populated
}
```

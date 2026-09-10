# Step


## Supported Types

### AssetNoteStep

```go
step := examplealias.NewStep(AssetNoteStep{/* values here */})
```

### AssetOutputStep

```go
step := examplealias.NewStep(AssetOutputStep{/* values here */})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch step.Type {
	case examplealias.StepTypeNote:
		// step.AssetNoteStep is populated
	case examplealias.StepTypeOutput:
		// step.AssetOutputStep is populated
	default:
		// Unknown type - use step.GetUnknownRaw() for raw JSON
}
```

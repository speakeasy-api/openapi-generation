# Status

## Example Usage

```csharp
using Speakeasy.OpenAPI;

var value = Status.InProgress;

// Open enum: use .Of() to create instances from custom string values
var custom = Status.Of("custom_value");
```


## Values

| Name             | Value            |
| ---------------- | ---------------- |
| `InProgress`     | in_progress      |
| `Completed`      | completed        |
| `Failed`         | failed           |
| `RequiresAction` | requires_action  |
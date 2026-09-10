# Status

## Example Usage

```java
import org.openapis.openapi.models.shared.Status;

Status value = Status.IN_PROGRESS;

// Open enum: use .of() to create instances from custom string values
Status custom = Status.of("custom_value");
```


## Values

| Name              | Value             |
| ----------------- | ----------------- |
| `IN_PROGRESS`     | in_progress       |
| `COMPLETED`       | completed         |
| `FAILED`          | failed            |
| `REQUIRES_ACTION` | requires_action   |
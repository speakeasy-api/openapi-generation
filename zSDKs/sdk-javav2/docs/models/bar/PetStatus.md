# PetStatus

Status of a pet in the bar namespace

## Example Usage

```java
import org.openapis.openapi.models.bar.PetStatus;

PetStatus value = PetStatus.AVAILABLE;

// Open enum: use .of() to create instances from custom string values
PetStatus custom = PetStatus.of("custom_value");
```


## Values

| Name          | Value         |
| ------------- | ------------- |
| `AVAILABLE`   | available     |
| `ADOPTED`     | adopted       |
| `PENDING`     | pending       |
| `UNAVAILABLE` | unavailable   |
# Enum

## Example Usage

```java
import org.openapis.openapi.models.shared.Enum;

Enum value = Enum.First;

// Open enum: use .of() to create instances from custom string values
Enum custom = Enum.of("custom_value");
```


## Values

| Name     | Value    |
| -------- | -------- |
| `First`  | First    |
| `Second` | Second   |
| `ThirdA` | Ex-aequo |
| `ThirdB` | Ex.aequo |
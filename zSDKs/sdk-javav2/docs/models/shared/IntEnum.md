# IntEnum

An integer enum property.

## Example Usage

```java
import org.openapis.openapi.models.shared.IntEnum;

IntEnum value = IntEnum.First;

// Open enum: use .of() to create instances from custom integer values
IntEnum custom = IntEnum.of(999);
```


## Values

| Name     | Value    |
| -------- | -------- |
| `First`  | 1        |
| `Second` | 2        |
| `Third`  | 3        |
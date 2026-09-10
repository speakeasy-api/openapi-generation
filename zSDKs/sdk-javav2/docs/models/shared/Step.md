# Step


## Supported Types

### Discriminator: `type`

| Value | Type |
| ----- | ---- |
| `"note"` | [AssetNoteStep](../../models/shared/AssetNoteStep.md) |
| `"output"` | [AssetOutputStep](../../models/shared/AssetOutputStep.md) |

### [`AssetNoteStep`](../../models/shared/AssetNoteStep.md)

Discriminator value: `"note"`

```java
Step value = AssetNoteStep.builder()
    .build();
```

### [`AssetOutputStep`](../../models/shared/AssetOutputStep.md)

Discriminator value: `"output"`

```java
Step value = AssetOutputStep.builder()
    .build();
```

## Consumption Patterns

### Java 11+ (Discriminator Switch)

```java
switch (value.type()) {
    case "note":
        // Handle note discriminator variant
        break;
    case "output":
        // Handle output discriminator variant
        break;
    default:
        // Handle unknown discriminator variant
}
```

### Java 16+ (Instanceof Pattern Matching)

```java
if (value instanceof AssetNoteStep assetNoteStep) {
    // Handle AssetNoteStep variant
} else if (value instanceof AssetOutputStep assetOutputStep) {
    // Handle AssetOutputStep variant
} else {
    // Handle unknown discriminator variant
}
```

### Java 21+ (Type Pattern Switch)

```java
switch (value) {
    case AssetNoteStep assetNoteStep -> {
        // Handle AssetNoteStep variant
    }
    case AssetOutputStep assetOutputStep -> {
        // Handle AssetOutputStep variant
    }
    default -> {
        // Handle unknown discriminator variant
    }
}
```

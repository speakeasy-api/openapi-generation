# Content


## Supported Types

### Discriminator: `type`

| Value | Type |
| ----- | ---- |
| `"text"` | [AssetTextBlock](../../models/shared/AssetTextBlock.md) |
| `"image"` | [AssetImageBlock](../../models/shared/AssetImageBlock.md) |

### [`AssetTextBlock`](../../models/shared/AssetTextBlock.md)

Discriminator value: `"text"`

```java
Content value = AssetTextBlock.builder()
    .build();
```

### [`AssetImageBlock`](../../models/shared/AssetImageBlock.md)

Discriminator value: `"image"`

```java
Content value = AssetImageBlock.builder()
    .build();
```

## Consumption Patterns

### Java 11+ (Discriminator Switch)

```java
switch (value.type()) {
    case "text":
        // Handle text discriminator variant
        break;
    case "image":
        // Handle image discriminator variant
        break;
    default:
        // Handle unknown discriminator variant
}
```

### Java 16+ (Instanceof Pattern Matching)

```java
if (value instanceof AssetTextBlock assetTextBlock) {
    // Handle AssetTextBlock variant
} else if (value instanceof AssetImageBlock assetImageBlock) {
    // Handle AssetImageBlock variant
} else {
    // Handle unknown discriminator variant
}
```

### Java 21+ (Type Pattern Switch)

```java
switch (value) {
    case AssetTextBlock assetTextBlock -> {
        // Handle AssetTextBlock variant
    }
    case AssetImageBlock assetImageBlock -> {
        // Handle AssetImageBlock variant
    }
    default -> {
        // Handle unknown discriminator variant
    }
}
```

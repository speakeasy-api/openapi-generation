# VariantRecord

A discriminated union with $-prefixed keys and x-speakeasy-discriminator overrides


## Supported Types

### Discriminator: `type`

| Value | Type |
| ----- | ---- |
| `"$text"` | [TextRecord](../../models/shared/TextRecord.md) |
| `"$image"` | [ImageRecord](../../models/shared/ImageRecord.md) |

### [`TextRecord`](../../models/shared/TextRecord.md)

Discriminator value: `"$text"`

```java
VariantRecord value = TextRecord.builder()
    .type("<value>")
    .text("sample text")
    .note("sample note")
    .build();
```

### [`ImageRecord`](../../models/shared/ImageRecord.md)

Discriminator value: `"$image"`

```java
VariantRecord value = ImageRecord.builder()
    .type("<value>")
    .imageId("image-001")
    .caption("sample caption")
    .build();
```

## Consumption Patterns

### Java 11+ (Discriminator Switch)

```java
switch (value.type()) {
    case "$text":
        // Handle $text discriminator variant
        break;
    case "$image":
        // Handle $image discriminator variant
        break;
    default:
        // Handle unknown discriminator variant
}
```

### Java 16+ (Instanceof Pattern Matching)

```java
if (value instanceof TextRecord textRecord) {
    // Handle TextRecord variant
} else if (value instanceof ImageRecord imageRecord) {
    // Handle ImageRecord variant
} else {
    // Handle unknown discriminator variant
}
```

### Java 21+ (Type Pattern Switch)

```java
switch (value) {
    case TextRecord textRecord -> {
        // Handle TextRecord variant
    }
    case ImageRecord imageRecord -> {
        // Handle ImageRecord variant
    }
    default -> {
        // Handle unknown discriminator variant
    }
}
```

# Shape

A discriminated union of shape types


## Supported Types

### Discriminator: `type`

| Value | Type |
| ----- | ---- |
| `"circle"` | [Circle](../../models/shared/Circle.md) |
| `"rectangle"` | [Rectangle](../../models/shared/Rectangle.md) |

### [`Circle`](../../models/shared/Circle.md)

Discriminator value: `"circle"`

```java
Shape value = Circle.builder()
    .type("<value>")
    .radius(9700.73)
    .build();
```

### [`Rectangle`](../../models/shared/Rectangle.md)

Discriminator value: `"rectangle"`

```java
Shape value = Rectangle.builder()
    .type("<value>")
    .width(6152.97)
    .height(6726.96)
    .build();
```

## Consumption Patterns

### Java 11+ (Discriminator Switch)

```java
switch (value.type()) {
    case "circle":
        // Handle circle discriminator variant
        break;
    case "rectangle":
        // Handle rectangle discriminator variant
        break;
    default:
        // Handle unknown discriminator variant
}
```

### Java 16+ (Instanceof Pattern Matching)

```java
if (value instanceof Circle circle) {
    // Handle Circle variant
} else if (value instanceof Rectangle rectangle) {
    // Handle Rectangle variant
} else {
    // Handle unknown discriminator variant
}
```

### Java 21+ (Type Pattern Switch)

```java
switch (value) {
    case Circle circle -> {
        // Handle Circle variant
    }
    case Rectangle rectangle -> {
        // Handle Rectangle variant
    }
    default -> {
        // Handle unknown discriminator variant
    }
}
```

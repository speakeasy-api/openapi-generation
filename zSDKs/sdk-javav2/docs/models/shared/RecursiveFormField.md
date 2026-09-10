# RecursiveFormField

A form field schema that can be recursive - the array type variant contains an array of RecursiveFormField items.


## Supported Types

### Discriminator: `type`

| Value | Type |
| ----- | ---- |
| `"text"` | [RecursiveFormFieldText](../../models/shared/RecursiveFormFieldText.md) |
| `"number"` | [RecursiveFormFieldNumber](../../models/shared/RecursiveFormFieldNumber.md) |
| `"array"` | [RecursiveFormFieldArray](../../models/shared/RecursiveFormFieldArray.md) |

### [`RecursiveFormFieldText`](../../models/shared/RecursiveFormFieldText.md)

Discriminator value: `"text"`

```java
RecursiveFormField value = RecursiveFormFieldText.builder()
    .label("<value>")
    .build();
```

### [`RecursiveFormFieldNumber`](../../models/shared/RecursiveFormFieldNumber.md)

Discriminator value: `"number"`

```java
RecursiveFormField value = RecursiveFormFieldNumber.builder()
    .label("<value>")
    .build();
```

### [`RecursiveFormFieldArray`](../../models/shared/RecursiveFormFieldArray.md)

Discriminator value: `"array"`

```java
RecursiveFormField value = RecursiveFormFieldArray.builder()
    .label("<value>")
    .itemType(RecursiveFormFieldNumber.builder()
        .label("<value>")
        .build())
    .build();
```

**Referred Types:** [RecursiveFormFieldNumber](../../models/shared/RecursiveFormFieldNumber.md)

## Consumption Patterns

### Java 11+ (Discriminator Switch)

```java
switch (value.type()) {
    case "text":
        // Handle text discriminator variant
        break;
    case "number":
        // Handle number discriminator variant
        break;
    case "array":
        // Handle array discriminator variant
        break;
    default:
        // Handle unknown discriminator variant
}
```

### Java 16+ (Instanceof Pattern Matching)

```java
if (value instanceof RecursiveFormFieldText recursiveFormFieldText) {
    // Handle RecursiveFormFieldText variant
} else if (value instanceof RecursiveFormFieldNumber recursiveFormFieldNumber) {
    // Handle RecursiveFormFieldNumber variant
} else if (value instanceof RecursiveFormFieldArray recursiveFormFieldArray) {
    // Handle RecursiveFormFieldArray variant
} else {
    // Handle unknown discriminator variant
}
```

### Java 21+ (Type Pattern Switch)

```java
switch (value) {
    case RecursiveFormFieldText recursiveFormFieldText -> {
        // Handle RecursiveFormFieldText variant
    }
    case RecursiveFormFieldNumber recursiveFormFieldNumber -> {
        // Handle RecursiveFormFieldNumber variant
    }
    case RecursiveFormFieldArray recursiveFormFieldArray -> {
        // Handle RecursiveFormFieldArray variant
    }
    default -> {
        // Handle unknown discriminator variant
    }
}
```

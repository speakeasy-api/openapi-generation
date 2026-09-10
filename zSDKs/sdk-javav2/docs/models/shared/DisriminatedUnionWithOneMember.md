# DisriminatedUnionWithOneMember


## Supported Types

### Discriminator: `type`

| Value | Type |
| ----- | ---- |
| `"type1"` | [ExhaustiveObject](../../models/shared/ExhaustiveObject.md) |

### [`ExhaustiveObject`](../../models/shared/ExhaustiveObject.md)

Discriminator value: `"type1"`

```java
DisriminatedUnionWithOneMember value = ExhaustiveObject.builder()
    .str("example")
    .bool(true)
    .integer(999999L)
    .int32(1)
    .num(1.1)
    .float32(4778.17f)
    .date(LocalDate.parse("2020-01-01"))
    .dateTime(OffsetDateTime.parse("2020-01-01T00:00:00Z"))
    .anything("<value>")
    .int32Enum(Int32Enum.SIXTY_NINE)
    .bigint(new BigInteger("978804"))
    .decimalStr(new BigDecimal("3470.92"))
    .obj(SimpleObject.builder()
        .str("example")
        .build())
    .map(Map.ofEntries(
    ))
    .arr(List.of())
    .any(Any.of("<value>"))
    .nullableStringEnum(NullableStringEnum.THIRD)
    .icon(Icon.TICK)
    .boolOpt(true)
    .intOptNull(999999L)
    .numOptNull(1.1)
    .intEnum(IntEnum.Third)
    .nullableIntEnum(NullableIntEnum.Third)
    .color(Color.GREEN)
    .heroWidth(HeroWidth.FOUR_HUNDRED_AND_EIGHTY)
    .build();
```

**Referred Types:**

- [Int32Enum](../../models/shared/Int32Enum.md)
- [SimpleObject](../../models/shared/SimpleObject.md)
- [Any](../../models/shared/Any.md)
- [NullableStringEnum](../../models/shared/NullableStringEnum.md)
- [Icon](../../models/shared/Icon.md)
- [IntEnum](../../models/shared/IntEnum.md)
- [NullableIntEnum](../../models/shared/NullableIntEnum.md)
- [Color](../../models/shared/Color.md)
- [HeroWidth](../../models/shared/HeroWidth.md)

## Consumption Patterns

### Java 11+ (Discriminator Switch)

```java
switch (value.type()) {
    case "type1":
        // Handle type1 discriminator variant
        break;
    default:
        // Handle unknown discriminator variant
}
```

### Java 16+ (Instanceof Pattern Matching)

```java
if (value instanceof ExhaustiveObject exhaustiveObject) {
    // Handle ExhaustiveObject variant
} else {
    // Handle unknown discriminator variant
}
```

### Java 21+ (Type Pattern Switch)

```java
switch (value) {
    case ExhaustiveObject exhaustiveObject -> {
        // Handle ExhaustiveObject variant
    }
    default -> {
        // Handle unknown discriminator variant
    }
}
```

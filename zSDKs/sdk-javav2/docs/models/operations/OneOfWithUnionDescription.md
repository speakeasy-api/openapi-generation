# OneOfWithUnionDescription

A union of two types.


## Supported Types

### [`ExhaustiveObject`](../../models/shared/ExhaustiveObject.md)

```java
OneOfWithUnionDescription value = OneOfWithUnionDescription.of(ExhaustiveObject.builder()
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
    .build());
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

### [`SimpleObject`](../../models/shared/SimpleObject.md)

```java
OneOfWithUnionDescription value = OneOfWithUnionDescription.of(SimpleObject.builder()
    .str("example")
    .build());
```

## Consumption Patterns

### Java 11+ (Accessor Methods)

```java
if (value.exhaustiveObject().isPresent()) {
    org.openapis.openapi.models.shared.ExhaustiveObject exhaustiveObjectValue = value.exhaustiveObject().get();
    // Handle exhaustiveObject variant
} else if (value.simpleObject().isPresent()) {
    org.openapis.openapi.models.shared.SimpleObject simpleObjectValue = value.simpleObject().get();
    // Handle simpleObject variant
} else if (value.asJson().isPresent()) {
    com.fasterxml.jackson.databind.JsonNode raw = value.asJson().get();
    // Handle unknown variant fallback
}
```

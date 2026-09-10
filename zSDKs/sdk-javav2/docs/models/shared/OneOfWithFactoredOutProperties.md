# OneOfWithFactoredOutProperties

A union of two types with factored out properties.


## Supported Types

### [`OneOfWithFactoredOutPropertiesExhaustiveObject`](../../models/shared/OneOfWithFactoredOutPropertiesExhaustiveObject.md)

```java
OneOfWithFactoredOutProperties value = OneOfWithFactoredOutProperties.of(OneOfWithFactoredOutPropertiesExhaustiveObject.builder()
    .str("example")
    .bool(true)
    .integer(999999L)
    .int32(1)
    .num(1.1)
    .float32(1866.41f)
    .date(LocalDate.parse("2020-01-01"))
    .dateTime(OffsetDateTime.parse("2020-01-01T00:00:00Z"))
    .anything("<value>")
    .int32Enum(OneOfWithFactoredOutPropertiesInt32Enum.SIXTY_NINE)
    .bigint(new BigInteger("817203"))
    .decimalStr(new BigDecimal("8930.15"))
    .obj(SimpleObject.builder()
        .str("example")
        .build())
    .map(Map.ofEntries(
    ))
    .arr(List.of(
        SimpleObject.builder()
            .str("example")
            .build()))
    .any(OneOfWithFactoredOutPropertiesAny.of(SimpleObject.builder()
        .str("example")
        .build()))
    .nullableStringEnum(OneOfWithFactoredOutPropertiesNullableStringEnum.SECOND)
    .icon(OneOfWithFactoredOutPropertiesIcon.TICK)
    .boolOpt(true)
    .intOptNull(999999L)
    .numOptNull(1.1)
    .intEnum(OneOfWithFactoredOutPropertiesIntEnum.Third)
    .nullableIntEnum(OneOfWithFactoredOutPropertiesNullableIntEnum.Third)
    .color(Color.GREEN)
    .heroWidth(OneOfWithFactoredOutPropertiesHeroWidth.FOUR_HUNDRED_AND_EIGHTY)
    .anExtraProperty("example")
    .build());
```

**Referred Types:**

- [OneOfWithFactoredOutPropertiesInt32Enum](../../models/shared/OneOfWithFactoredOutPropertiesInt32Enum.md)
- [SimpleObject](../../models/shared/SimpleObject.md)
- [OneOfWithFactoredOutPropertiesAny](../../models/shared/OneOfWithFactoredOutPropertiesAny.md)
- [OneOfWithFactoredOutPropertiesNullableStringEnum](../../models/shared/OneOfWithFactoredOutPropertiesNullableStringEnum.md)
- [OneOfWithFactoredOutPropertiesIcon](../../models/shared/OneOfWithFactoredOutPropertiesIcon.md)
- [OneOfWithFactoredOutPropertiesIntEnum](../../models/shared/OneOfWithFactoredOutPropertiesIntEnum.md)
- [OneOfWithFactoredOutPropertiesNullableIntEnum](../../models/shared/OneOfWithFactoredOutPropertiesNullableIntEnum.md)
- [Color](../../models/shared/Color.md)
- [OneOfWithFactoredOutPropertiesHeroWidth](../../models/shared/OneOfWithFactoredOutPropertiesHeroWidth.md)

### [`OneOfWithFactoredOutPropertiesSimpleObject`](../../models/shared/OneOfWithFactoredOutPropertiesSimpleObject.md)

```java
OneOfWithFactoredOutProperties value = OneOfWithFactoredOutProperties.of(OneOfWithFactoredOutPropertiesSimpleObject.builder()
    .str("example")
    .anExtraProperty("example")
    .build());
```

## Consumption Patterns

### Java 11+ (Accessor Methods)

```java
if (value.oneOfWithFactoredOutPropertiesExhaustiveObject().isPresent()) {
    org.openapis.openapi.models.shared.OneOfWithFactoredOutPropertiesExhaustiveObject oneOfWithFactoredOutPropertiesExhaustiveObjectValue = value.oneOfWithFactoredOutPropertiesExhaustiveObject().get();
    // Handle oneOfWithFactoredOutPropertiesExhaustiveObject variant
} else if (value.oneOfWithFactoredOutPropertiesSimpleObject().isPresent()) {
    org.openapis.openapi.models.shared.OneOfWithFactoredOutPropertiesSimpleObject oneOfWithFactoredOutPropertiesSimpleObjectValue = value.oneOfWithFactoredOutPropertiesSimpleObject().get();
    // Handle oneOfWithFactoredOutPropertiesSimpleObject variant
} else if (value.asJson().isPresent()) {
    com.fasterxml.jackson.databind.JsonNode raw = value.asJson().get();
    // Handle unknown variant fallback
}
```

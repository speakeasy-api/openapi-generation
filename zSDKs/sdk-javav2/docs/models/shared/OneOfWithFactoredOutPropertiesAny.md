# OneOfWithFactoredOutPropertiesAny


## Supported Types

### [`SimpleObject`](../../models/shared/SimpleObject.md)

```java
OneOfWithFactoredOutPropertiesAny value = OneOfWithFactoredOutPropertiesAny.of(SimpleObject.builder()
    .str("example")
    .build());
```

### `String`

```java
OneOfWithFactoredOutPropertiesAny value = OneOfWithFactoredOutPropertiesAny.of("<value>");
```

## Consumption Patterns

### Java 11+ (Accessor Methods)

```java
if (value.simpleObject().isPresent()) {
    org.openapis.openapi.models.shared.SimpleObject simpleObjectValue = value.simpleObject().get();
    // Handle simpleObject variant
} else if (value.string().isPresent()) {
    java.lang.String stringValue = value.string().get();
    // Handle string variant
} else if (value.asJson().isPresent()) {
    com.fasterxml.jackson.databind.JsonNode raw = value.asJson().get();
    // Handle unknown variant fallback
}
```

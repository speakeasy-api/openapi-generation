# CircularUnion


## Supported Types

### `Map<String, CircularUnion>`

```java
CircularUnion value = CircularUnion.of(Map.ofEntries(
));
```

### `String`

```java
CircularUnion value = CircularUnion.of("<value>");
```

### `long`

```java
CircularUnion value = CircularUnion.of(725973L);
```

### `boolean`

```java
CircularUnion value = CircularUnion.of(true);
```

### `List<CircularUnion>`

```java
CircularUnion value = CircularUnion.of(List.of(
    CircularUnion.of(47.37)));
```

**Referred Types:** [CircularUnion](../../models/shared/CircularUnion.md)

### `double`

```java
CircularUnion value = CircularUnion.of(2785.03);
```

## Consumption Patterns

### Java 11+ (Accessor Methods)

```java
if (value.mapOfCircularUnion().isPresent()) {
    java.util.Map<java.lang.String, org.openapis.openapi.models.shared.CircularUnion> mapOfCircularUnionValue = value.mapOfCircularUnion().get();
    // Handle mapOfCircularUnion variant
} else if (value.string().isPresent()) {
    java.lang.String stringValue = value.string().get();
    // Handle string variant
} else if (value.asLong().isPresent()) {
    java.lang.Long longValue = value.asLong().get();
    // Handle asLong variant
} else if (value.asBoolean().isPresent()) {
    java.lang.Boolean booleanValue = value.asBoolean().get();
    // Handle asBoolean variant
} else if (value.arrayOfCircularUnion().isPresent()) {
    java.util.List<org.openapis.openapi.models.shared.CircularUnion> arrayOfCircularUnionValue = value.arrayOfCircularUnion().get();
    // Handle arrayOfCircularUnion variant
} else if (value.asDouble().isPresent()) {
    java.lang.Double doubleValue = value.asDouble().get();
    // Handle asDouble variant
} else if (value.asJson().isPresent()) {
    com.fasterxml.jackson.databind.JsonNode raw = value.asJson().get();
    // Handle unknown variant fallback
}
```

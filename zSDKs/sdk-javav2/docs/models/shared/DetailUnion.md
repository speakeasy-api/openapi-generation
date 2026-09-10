# DetailUnion

Contains parameter or domain specific information related to the error and why it occurred.


## Supported Types

### `String`

```java
DetailUnion value = DetailUnion.of("Missing property foobar");
```

### [`Detail`](../../models/shared/Detail.md)

```java
DetailUnion value = DetailUnion.of(Detail.builder()
    .build());
```

## Consumption Patterns

### Java 11+ (Accessor Methods)

```java
if (value.string().isPresent()) {
    java.lang.String stringValue = value.string().get();
    // Handle string variant
} else if (value.detail().isPresent()) {
    org.openapis.openapi.models.shared.Detail detailValue = value.detail().get();
    // Handle detail variant
} else if (value.asJson().isPresent()) {
    com.fasterxml.jackson.databind.JsonNode raw = value.asJson().get();
    // Handle unknown variant fallback
}
```

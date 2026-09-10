# SomeUnion

A union demonstrating title and x-speakeasy-name-override on primitive types


## Supported Types

### `String`

```java
SomeUnion value = SomeUnion.of("<value>");
```

### [`MyObject`](../../models/shared/MyObject.md)

```java
SomeUnion value = SomeUnion.of(MyObject.builder()
    .build());
```

### `boolean`

```java
SomeUnion value = SomeUnion.of(true);
```

## Consumption Patterns

### Java 11+ (Accessor Methods)

```java
if (value.myString().isPresent()) {
    java.lang.String myStringValue = value.myString().get();
    // Handle myString variant
} else if (value.myObject().isPresent()) {
    org.openapis.openapi.models.shared.MyObject myObjectValue = value.myObject().get();
    // Handle myObject variant
} else if (value.asFoo().isPresent()) {
    java.lang.Boolean fooValue = value.asFoo().get();
    // Handle asFoo variant
} else if (value.asJson().isPresent()) {
    com.fasterxml.jackson.databind.JsonNode raw = value.asJson().get();
    // Handle unknown variant fallback
}
```

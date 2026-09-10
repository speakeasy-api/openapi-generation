# Vehicle

A non-discriminated union of vehicle types in the bar namespace


## Supported Types

### [`Car`](../../models/bar/Car.md)

```java
Vehicle value = Vehicle.of(Car.builder()
    .engineType("V8")
    .build());
```

### [`Bike`](../../models/bar/Bike.md)

```java
Vehicle value = Vehicle.of(Bike.builder()
    .hasPedals(true)
    .build());
```

## Consumption Patterns

### Java 11+ (Accessor Methods)

```java
if (value.car().isPresent()) {
    org.openapis.openapi.models.bar.Car carValue = value.car().get();
    // Handle car variant
} else if (value.bike().isPresent()) {
    org.openapis.openapi.models.bar.Bike bikeValue = value.bike().get();
    // Handle bike variant
} else if (value.asJson().isPresent()) {
    com.fasterxml.jackson.databind.JsonNode raw = value.asJson().get();
    // Handle unknown variant fallback
}
```

# Animal

A discriminated union of animal types in the foo namespace


## Supported Types

### Discriminator: `animalType`

| Value | Type |
| ----- | ---- |
| `"dog"` | [Dog](../../models/foo/Dog.md) |
| `"cat"` | [Cat](../../models/foo/Cat.md) |

### [`Dog`](../../models/foo/Dog.md)

Discriminator value: `"dog"`

```java
Animal value = Dog.builder()
    .animalType(DogAnimalType.DOG)
    .breed("Labrador")
    .barkVolume(8L)
    .build();
```

**Referred Types:** [DogAnimalType](../../models/foo/DogAnimalType.md)

### [`Cat`](../../models/foo/Cat.md)

Discriminator value: `"cat"`

```java
Animal value = Cat.builder()
    .animalType(CatAnimalType.CAT)
    .furLength("short")
    .meowFrequency(5L)
    .build();
```

**Referred Types:** [CatAnimalType](../../models/foo/CatAnimalType.md)

## Consumption Patterns

### Java 11+ (Discriminator Switch)

```java
switch (value.animalType()) {
    case "dog":
        // Handle dog discriminator variant
        break;
    case "cat":
        // Handle cat discriminator variant
        break;
    default:
        // Handle unknown discriminator variant
}
```

### Java 16+ (Instanceof Pattern Matching)

```java
if (value instanceof Dog dog) {
    // Handle Dog variant
} else if (value instanceof Cat cat) {
    // Handle Cat variant
} else {
    // Handle unknown discriminator variant
}
```

### Java 21+ (Type Pattern Switch)

```java
switch (value) {
    case Dog dog -> {
        // Handle Dog variant
    }
    case Cat cat -> {
        // Handle Cat variant
    }
    default -> {
        // Handle unknown discriminator variant
    }
}
```

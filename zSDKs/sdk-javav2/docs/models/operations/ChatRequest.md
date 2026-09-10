# ChatRequest


## Supported Types

### [`ChatModelRequest`](../../models/shared/ChatModelRequest.md)

```java
ChatRequest value = ChatRequest.of(ChatModelRequest.builder()
    .prompt("What is the largest city in the world?")
    .build());
```

### [`ChatAgentRequest`](../../models/shared/ChatAgentRequest.md)

```java
ChatRequest value = ChatRequest.of(ChatAgentRequest.builder()
    .agent("<value>")
    .prompt("<value>")
    .build());
```

## Consumption Patterns

### Java 11+ (Accessor Methods)

```java
if (value.chatModelRequest().isPresent()) {
    org.openapis.openapi.models.shared.ChatModelRequest chatModelRequestValue = value.chatModelRequest().get();
    // Handle chatModelRequest variant
} else if (value.chatAgentRequest().isPresent()) {
    org.openapis.openapi.models.shared.ChatAgentRequest chatAgentRequestValue = value.chatAgentRequest().get();
    // Handle chatAgentRequest variant
} else if (value.asJson().isPresent()) {
    com.fasterxml.jackson.databind.JsonNode raw = value.asJson().get();
    // Handle unknown variant fallback
}
```

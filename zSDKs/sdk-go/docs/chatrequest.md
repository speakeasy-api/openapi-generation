# ChatRequest


## Supported Types

### ChatModelRequest

```go
chatRequest := examplealias.NewChatRequest(ChatModelRequest{/* values here */})
```

### ChatAgentRequest

```go
chatRequest := examplealias.NewChatRequest(ChatAgentRequest{/* values here */})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch chatRequest.Type {
	case examplealias.ChatRequestTypeChatModelRequest:
		// chatRequest.ChatModelRequest is populated
	case examplealias.ChatRequestTypeChatAgentRequest:
		// chatRequest.ChatAgentRequest is populated
}
```

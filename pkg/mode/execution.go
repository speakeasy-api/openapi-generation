package mode

import (
	"context"
)

type contextKey string

const ValidationContextKey = contextKey("speakeasy.execution_context")
const EmbeddedContextKey = contextKey("speakeasy.is_embedded")

func SetSpeakeasyExecutionContextValidation(ctx context.Context) context.Context {
	return context.WithValue(ctx, ValidationContextKey, true)
}

func SetSpeakeasyExecutionContextEmbedded(ctx context.Context) context.Context {
	return context.WithValue(ctx, EmbeddedContextKey, true)
}

func IsSpeakeasyExecutionContextValidation(ctx context.Context) bool {
	// If not set, return false
	if ctx.Value(ValidationContextKey) == nil {
		return false
	}

	return true
}

func IsSpeakeasyExecutionContextEmbedded(ctx context.Context) bool {
	// If not set, return false
	if ctx.Value(EmbeddedContextKey) == nil {
		return false
	}

	return true
}
